wrk.method="POST"
wrk.headers["Content-Type"] = "application/json"
-- 这个要改为你的注册的数据
wrk.body='{"email": "110@qq.com", "password": "123456789", "confirmPassword": "123456789"}'