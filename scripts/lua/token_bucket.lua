-- token_bucket.lua
-- Atomic token bucket rate limiter for Redis
-- Load with SCRIPT LOAD and invoke via EVALSHA for script caching.
--
-- KEYS[1] = state key (e.g. ratelimit:token_bucket:user:123)
-- ARGV[1] = capacity     (int,   max tokens)
-- ARGV[2] = refill_rate  (float, tokens per second; 0 = no refill)
-- ARGV[3] = requested    (int,   tokens to consume; 0 = peek without consuming)
-- ARGV[4] = ttl_sec      (int,   key expiration in seconds)
--
-- Returns: {allowed, remaining, retry_after_ms, reset_at_ms}
--   allowed        1 = request allowed, 0 = denied
--   remaining      tokens left after this request (floor)
--   retry_after_ms ms until enough tokens for `requested` (0 if allowed)
--   reset_at_ms    ms until bucket is full (epoch ms); 2147483647 if no refill

local key        = KEYS[1]
local capacity   = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local requested  = tonumber(ARGV[3])
local ttl_sec    = tonumber(ARGV[4])

-- ── Input guards ────────────────────────────────────────────────────────────

if capacity == nil or capacity <= 0 then
    return redis.error_reply("ERR capacity must be a positive number")
end
if refill_rate == nil or refill_rate < 0 then
    return redis.error_reply("ERR refill_rate must be >= 0")
end
if requested == nil or requested < 0 then
    return redis.error_reply("ERR requested must be >= 0")
end
if ttl_sec == nil or ttl_sec <= 0 then
    return redis.error_reply("ERR ttl_sec must be a positive number")
end

local no_refill = (refill_rate == 0)

-- ── Authoritative timestamp from Redis (avoids caller clock skew) ──────────

local time_result = redis.call('TIME')
local now_sec  = tonumber(time_result[1])
local now_usec = tonumber(time_result[2])
local now_ms   = now_sec * 1000 + math.floor(now_usec / 1000)

-- ── Peek shortcut (requested == 0) ─────────────────────────────────────────
-- Read current token count without mutating state.

if requested == 0 then
    local peek = redis.call('HGET', key, 'tokens')
    local cur  = tonumber(peek) or capacity
    return {1, math.floor(cur), 0, now_ms}
end

-- ── Load existing state ─────────────────────────────────────────────────────

local state    = redis.call('HMGET', key, 'tokens', 'last_refill_sec', 'last_refill_usec')
local tokens   = tonumber(state[1])
local last_sec  = tonumber(state[2])
local last_usec = tonumber(state[3])

-- ── Initialise or refill ────────────────────────────────────────────────────

if tokens == nil then
    -- New key: start at full capacity.
    tokens    = capacity
    last_sec  = now_sec
    last_usec = now_usec
else
    if not no_refill then
        -- Clamp elapsed to 0 to survive Redis failover / clock drift backwards.
        local delta_sec  = (now_sec  - last_sec)
        local delta_usec = (now_usec - last_usec)
        local elapsed    = math.max(0, delta_sec + delta_usec / 1000000.0)
        if elapsed > 0 then
            tokens = math.min(tokens + elapsed * refill_rate, capacity)
        end
    end
    last_sec  = now_sec
    last_usec = now_usec
end

-- ── Decision ────────────────────────────────────────────────────────────────

local allowed        = 0
local remaining      = 0
local retry_after_ms = 0
local reset_at_ms    = 0

if tokens >= requested then
    tokens = tokens - requested
    allowed = 1
    remaining = math.floor(tokens)
    -- reset_at_ms = time the bucket will be full again (or now if already full).
    if not no_refill and tokens < capacity then
        local deficit = capacity - tokens
        reset_at_ms   = now_ms + math.ceil(deficit / refill_rate * 1000)
    else
        reset_at_ms = now_ms
    end
else
    -- Denied: tokens is not mutated so the state is unchanged.
    remaining = 0
    if no_refill then
        retry_after_ms = 2147483647
        reset_at_ms    = 2147483647
    else
        -- retry_after_ms: wait until THIS request can be served.
        local need        = requested - tokens
        retry_after_ms    = math.ceil(need / refill_rate * 1000)
        -- reset_at_ms: wait until bucket is completely full.
        local to_full     = capacity - tokens
        reset_at_ms       = now_ms + math.ceil(to_full / refill_rate * 1000)
    end
end

-- ── Persist state ───────────────────────────────────────────────────────────
-- tokens is stored as a float string so fractional tokens survive across calls.
-- Reads always use tonumber() which handles float strings correctly.

redis.call('HSET', key,
    'tokens',           tostring(tokens),
    'last_refill_sec',  tostring(last_sec),
    'last_refill_usec', tostring(last_usec))
redis.call('EXPIRE', key, ttl_sec)

return {allowed, remaining, math.floor(retry_after_ms), math.floor(reset_at_ms)}