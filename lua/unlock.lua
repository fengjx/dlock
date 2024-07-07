-- 1: 删除锁成功返回
-- -1: 锁不存在或者已过期
-- 0: 锁存在，但是 value 不相等，不允许删除
local val = redis.call("GET", KEYS[1])
if val == ARGV[1] then
    -- 已存在，则删除
    return redis.call("DEL", KEYS[1])
elseif val == nil then
    -- 不存在，则返回-1
    return -1
else
    -- key 存在，但是 value 不相等
    return 0
end