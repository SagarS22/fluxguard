-- sliding_window.lua
-- Atomic sliding-window-log rate limiter for Redis.
--
-- Load with SCRIPT LOAD and invoke via EVALSHA for script caching.
--
-- IMPORTANT (Redis Cluster): this script uses two keys. In cluster mode both keys
-- must hash to the same slot. Use a shared hash tag, for example:
--   KEYS[1] = ratelimit:sliding_window:user:{dXNlcjoxMjM}
--   KEYS[2] = ratelimit:sliding_window:user:{dXNlcjoxMjM}:seq
--
-- KEYS[1] = request log ZSET key
-- KEYS[2] = sequence key for unique member generation
--
-- ARGV[1] = limit       (int, max requests allowed in the window)
-- ARGV[2] = window_ms   (int, rolling window size in milliseconds)
-- ARGV[3] = requested   (int, units to consume; 0 = peek without mutation)
-- ARGV[4] = ttl_sec     (int, key expiration in seconds; effective TTL is at least the window)
--
-- Returns: {allowed, remaining, retry_after_ms, reset_at_ms}
--   allowed        1 = request allowed, 0 = denied
--   remaining      requests left after this check
--   retry_after_ms ms until enough capacity exists for `requested` (0 if allowed)
--   reset_at_ms    epoch ms when all currently counted requests have expired

local log_key    = KEYS[1]
local seq_key    = KEYS[2]
local limit      = tonumber(ARGV[1])
local window_ms  = tonumber(ARGV[2])
local requested  = tonumber(ARGV[3])
local ttl_sec    = tonumber(ARGV[4])

if limit == nil or limit <= 0 or math.floor(limit) ~= limit then
    return redis.error_reply('ERR limit must be a positive integer')
end
if window_ms == nil or window_ms <= 0 or math.floor(window_ms) ~= window_ms then
    return redis.error_reply('ERR window_ms must be a positive integer')
end
if requested == nil or requested < 0 or math.floor(requested) ~= requested then
    return redis.error_reply('ERR requested must be a non-negative integer')
end
if requested > limit then
    return redis.error_reply('ERR requested must be <= limit')
end
if ttl_sec == nil or ttl_sec <= 0 or math.floor(ttl_sec) ~= ttl_sec then
    return redis.error_reply('ERR ttl_sec must be a positive integer')
end

local ttl_ms = math.max(ttl_sec * 1000, window_ms)

-- Redis TIME is authoritative and avoids caller clock skew.
local time_result = redis.call('TIME')
local now_sec     = tonumber(time_result[1])
local now_usec    = tonumber(time_result[2])
local now_ms      = now_sec * 1000 + math.floor(now_usec / 1000)
local trim_before = now_ms - window_ms

-- Keep only requests strictly inside the sliding window: (now - window_ms, now].
redis.call('ZREMRANGEBYSCORE', log_key, '-inf', trim_before)

local current_count = redis.call('ZCARD', log_key)

local function compute_full_reset_at_ms(count)
    if count == 0 then
        return now_ms
    end

    local latest = redis.call('ZREVRANGE', log_key, 0, 0, 'WITHSCORES')
    local latest_score = tonumber(latest[2])
    if latest_score == nil then
        return now_ms
    end

    return latest_score + window_ms
end

if requested == 0 then
    return {1, math.max(0, limit - current_count), 0, compute_full_reset_at_ms(current_count)}
end

if current_count + requested <= limit then
    local seq_end = redis.call('INCRBY', seq_key, requested)
    local seq_start = seq_end - requested + 1
    local zadd_args = {log_key}

    for seq = seq_start, seq_end do
        zadd_args[#zadd_args + 1] = now_ms
        zadd_args[#zadd_args + 1] = tostring(now_sec) .. ':' .. tostring(now_usec) .. ':' .. tostring(seq)
    end

    redis.call('ZADD', unpack(zadd_args))
    redis.call('PEXPIRE', log_key, ttl_ms)
    redis.call('PEXPIRE', seq_key, ttl_ms)

    local new_count = current_count + requested
    return {1, limit - new_count, 0, now_ms + window_ms}
end

local entries_to_expire = current_count + requested - limit
local retry_after_ms = window_ms

local blocking_entry = redis.call('ZRANGE', log_key, entries_to_expire - 1, entries_to_expire - 1, 'WITHSCORES')
local blocking_score = tonumber(blocking_entry[2])
if blocking_score ~= nil then
    retry_after_ms = math.max(0, (blocking_score + window_ms) - now_ms)
end

return {0, math.max(0, limit - current_count), retry_after_ms, compute_full_reset_at_ms(current_count)}
