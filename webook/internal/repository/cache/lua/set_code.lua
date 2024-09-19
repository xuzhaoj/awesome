-- 发送到的 key, 也就是 code:业务:手机号
local key = KEYS[1]
-- 使用次数, 也就是验证次数.记录了验证几次
-- phone_code:login:155xxxxxxx:cnt
local cntKey = key..":cnt"
-- 验证码的值
local val = ARGV[1]
-- 验证码的有效时间是十分钟, 600 秒,过期时间
local ttl = tonumber(redis.call("ttl", key))

if ttl == -1 then
    -- key 存在，但是没有过期时间
    -- 系统错误, 你的同事手误, 手动设置了这个 key, 但是没给过期时间
    --redis.log(redis.LOG_NOTICE, "TTL is -1: Key exists but has no expiration.")
    return -2
-- 验证码一分钟的有效期已经到了，重新发送或者key不存在
elseif ttl == -2 or ttl < 540 then
    -- 完美，符合预期
    -- 540 = 600-60 九分钟
    --设置新的key
    redis.call("set", key, val)
    --设置新的过期时间
    redis.call("expire", key, 600)
    --跟踪使用次数
    redis.call("set", cntKey, 3)
    --10分钟没有用完3次验证机会过期后会自动删除
    redis.call("expire", cntKey, 600)
    --redis.log(redis.LOG_NOTICE, "TTL is -2 or less than 540: Key and count reset.")
    return 0
else
    -- 发送太频繁
    --redis.log(redis.LOG_NOTICE, "TTL is greater than or equal to 540: Requests are too frequent.")
    return -1
end
