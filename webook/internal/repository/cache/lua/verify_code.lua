local key = KEYS[1]
-- 用户的输入 code
local expectedCode = ARGV[1]
local code = redis.call("get", key)
local cntKey = key..":cnt"
-- 转成一个数字
local cnt = tonumber(redis.call("get", cntKey))

if cnt <= 0 then
    -- 说明，用户一直输错
    return -1
elseif expectedCode == code then
    -- 输入对了,设置为-1表示不可以再用了
    redis.call("set",cntKey,-1)
    return 0
else
    -- 用户手抖，输错了
    -- 可验证次数减一
    redis.call("decr",cntKey)
    return -2
end