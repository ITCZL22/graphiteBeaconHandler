## graphiteBeaconHandler
### Features
- 对不同graphite_beacon的alert支持不同的报警分组
- 相同前缀的alert使用同一组报警
- 自定义报警方式(目前支持mail & slack)
- 不仅针对graphite-beacon报警，还可以直接使用http请求调用

### Service Config Example
[conf.json](./conf.json)
### Graphite_beacon Config Example 
[graphite_beacon.json](./graphite_beacon.json)

### Http Request
- curl -X POST http://itczl.com:8765 -d "level=warning&desc=describe&alert=ITCZL:XXXX&noemail=1"
- level: 报警级别
- desc: 报警内容
- alert: 报警分组
- noemail: 可选参数,noemail=1不启用邮件通知,优先级大于配置文件

### reference
- [graphite](https://github.com/klen/graphite-beacon)
- [graphite-beacon](https://github.com/graphite-project/graphite-web)
