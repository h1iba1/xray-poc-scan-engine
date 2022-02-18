## 📝 日志库-logo

## 需求分析

1. 支持往不同的地方输出日志

2. 日志分级别
   - Debug
   - Trace
   - Info
   - Warning
   - Error
   - Fatal
   
3. 日志要支持开关控制

4. 完整的日志记录要包含时间、行号、文件名、日志级别、日志信息

5. 打印日志可以定义输出格式，至少有text和json两种格式

6. 日志文件要切割

   - 按文件大小切割

     ```go
     1. 关闭当前文件
     2. 备份一个 rename
     3. 打开一个新的日志文件
     4. 将打开的文件赋值给 fl.FileObj
     ```
     
   - 按日期切割
   
     ```go
     1. 在日志结构体中设置一个字段记录上一次切割的小时数
     2. 再写日志之前检查一下当前时间的小时数和保存的是否一致，不一致就要切割
     ```
   
   - 设置日志最大保留时长
   
     ```go
     1. 在日志结构体中设置一个字段记录最大文件保留时长
     2. 定时扫描所有日志文件，若日志文件时间早于保留最早时间，则删除
     ```
   
   - 设置日志文件最大保留个数
   
     ```
     1. 在日志结构体中设置一个字段记录最大保留文件个数
     2. 定时扫描所有日志文件。 若日志文件数量超出指定个数时，保留指定个数最新文件，其余文件删除
     ```
7. 性能优化：异步打印日志
## 更新日志

- 2020.06.15 增加异步打印日志功能和日志同时输出console和file功能
- 2020.06.16 增加自定义日志文件最大保留时长功能
- 2020.06.17 增加自定义日志文件最大保留个数功能，生成文件Logger时可以传不定长参数
- 2020.06.18 增加日志输出Json格式，优化日志打印效率
- 2020.06.24 修改日志输入文件参数解析框架，提高参数传值和解析效率
- 2020.07.05 实现自定义级别打印文件行信息功能