1. 表结构设计
   - 使用 `user_id` 而不使用 自增`id` 作为用户标识
     - `user_id` 作为用户标识，可以更好的保护用户隐私
     - 使用分库、分表时，自增`id` 会有问题，而 `user_id` 不会
2. 分布式ID生成
   1. 分布式ID特点
      - 全局唯一：不能出现有重复的ID标识，这是基本要求
      - 趋势递增：确保生成的ID在一定程度上是递增的，这样生成的ID在索引上有更好的性能
      - 高可用性：确保在任何时候都能生成正确的ID
      - 高性能性：在高并发的环境下依然表现良好
   2. 雪花算法(SnowFlake)
      - 64位整数，分别表示
        - 1位：占用1bit，其值始终是0，没有实际作用
        - 41位：占用41bit，时间戳，毫秒级(约69年)
        - 10位：工作机器标识ID，高位5bit数据中心ID，低位5bit工作ID(最多支持32个数据中心，每个数据中心最多支持32个节点)
        - 12位：序列号，毫秒内的计数(同一毫秒可以生成4096个ID，算上机器ID：1024 * 4096)
      - 优点
        - 高性能高可用，生成ID时不依赖数据库，完全在内存中生成
        - 高吞吐，每秒钟能够生成几百万个自增ID
        - ID自增，存入数据库中，效率高
      - 缺点
        - 依赖服务器时间，服务器时间回拨时可能会生成重复 id。
          - **原因：**
            1. 人为原因，把系统环境的时间改了
            2. 有时候不同的机器上需要同步时间，可能不同机器之间存在误差，那么可能会出现时间回拨问题。
          - **解决方案：**
            1. 算法中可通过记录最后一个生成 id 时的时间戳来解决，每次生成 id 之前比较当前服务器时钟是否被回拨，避免生成重复 id。
            2. 可参考美团百度
        - 在单机上，生成的ID是递增的，但在多台机器上，只能大致保持递增趋势，并不能严格保证递增
          - **原因：**
            1. 这是因为多台机器之间的时钟不一定完全同步。
        - 雪花算法依赖于时间的一致性，如果发生时间回拨，可能会导致问题。
          - **解决方案：**
            1. 通常会使用拓展位来扩展时间戳的位数。
      - 参考链接：
        - [一文读懂“Snowflake（雪花）”算法](https://cloud.tencent.com/developer/article/2364487)
3. 用户认证
    - 原因：`HTTP`是一个无状态的协议，一次请求结束后，下次在发送服务器就不知道这个请求是谁发来的了（同一个IP不代表同一个用户）
    - 方案：
        1. **`Cookie - Session`模式会话管理流程**
            - 客户端使用用户名、密码进行认证
            - 服务端验证用户名、密码正确后生成并存储`Session`信息，将`Session`信息存储在`Cookie`中，返回给客户端
            - 客户端访问需要认证的接口时，在`Cookie`中携带`Session`信息
            - 服务端通过`SessionID`查找`Session`信息进行鉴权，返回给客户端需要的信息
        2. **`Token`模式会话管理流程**
            - 客户端使用用户名、密码进行认证
            - 服务端验证用户名、密码正确后生成`Token`信息，返回给客户端
            - 客户端村存储`Token`信息，访问需要认证的接口时，在`URL`参数或者`HTTP Header`中加入`Token`信息
            - 服务端需要通过解码`Token`信息来进行鉴权，返回给客户端需要的信息
        3. `Cookie - Session`模式缺点
            1. 服务端需要存储`Session`信息，并且由于`Session`信息需要经常快速查找，通常存储在内存或者内存数据库中，当同时在线用户较多时会大量占用服务器资源
            2. 当需要扩展时，创建`Session`的服务器可能不是验证`Session`的服务器，所以还需要将所有的`Session`信息单独存储并共享
            3. 由于客户端使用`Cookie`存储`Session`信息，在跨域场景下需要进行兼容性处理，同时这种方式也难以防范`CSRF`攻击
        4. `Token`模式优点
            1. 服务端不需要存储和用户鉴权有关的信息，鉴权信息会被加密到`Token`中，服务端只需要解码`Token`信息即可
            2. 避免了共享`Session`信息导致的不易扩展问题
            3. 不需要依赖`Cookie`，有效避免`Cookie`带来的`CSRF`攻击 
            4. 使用`CORS`可以快速解决跨域问题
        5. `JWT(JSON Web Token)`
            - `JWT`是一种开放标准(RFC 7519)，本身并没有定义任何技术实现，只是定义了一种基于`Token`的会话管理的规则
            - `JWT`由三部分组成：`Header`、`Payload`、`Signature`
                - `Header`：包含`Token`的元数据，通常包含`Token`的类型和加密算法
                - `Payload`：包含`Token`的主要内容，通常包含用户信息、权限信息等(此部分不加密，只是`Base64`编码)
                    1. `iss(issuser)`: 签发人
                    2. `exp(expiration time)`: 过期时间
                    3. `sub(subject)`: 主题
                    4. `aud(audience)`: 受众
                    5. `nbf(not before)`: 生效时间
                    6. `iat(issued at)`: 签发时间
                    7. `jti(JWT ID)`: 编号
                - `Signature`：`Header`和`Payload`的签名，防止数据被篡改
            - 公式：
                - `Token = Base64(Header) + '.' + Base64(Payload) + '.' + Signature`
        6. `JWT`的优缺点
            - 优点
                1. 无状态：`JWT`本身包含了用户信息，服务端不需要存储用户信息，只需要解码`Token`即可
                2. 跨域：`JWT`可以通过`CORS`解决跨域问题
                   cc3. 不再需要存储`Session`信息，避免了`Session`信息共享导致的不易扩展问题
            - 缺点
                1. 无法撤销：`JWT`一旦签发，就无法撤销(在服务端废止)，除非等待`Token`过期
                2. `Payload`信息不加密：`JWT`的`Payload`部分是经过`Base64`编码的，不是加密的，所以不要在`Payload`中存储敏感信息
        7. 基于`JWT`的`Refresh Token`机制
            - `Refresh Token`的有效期通常比`Token`长，一般为`7`天或者`30`天'
            - `Refresh Token`的作用是在`Token`过期后，通过`Refresh Token`获取新的`Token`，而不需要重新登录
            - 当`Refresh Token`过期后，需要重新登录
            - 会话管理流程:
                - 客户端使用用户名、密码进行认证
                - 服务端生成有效时间较短的`Access Token`（例如10min）和有效时间较长的`Refresh Token`（例如7天）
                - 客户端访问需要认证的接口时，携带`Access Token`
                - 如果`Access Token`没有过期，服务器鉴权后返回客户端所需要的数据
                - 如果携带`Access Token`访问需要认证的接口时鉴权失败(例如返回`401`错误)，客户端使用`Refresh Token`获取新的`Access Token`
                - 如果`Refresh Token`没有过期，服务端向客户端发送新的`Access Token`
                - 客户端使用新的`Access Token`访问需要认证的接口



​    
