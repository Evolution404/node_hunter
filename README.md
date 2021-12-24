# NODE HUNTER
**以太坊网络探测工具**
## 编译运行
```sh
# 克隆项目
git clone https://github.com/Evolution404/node_hunter.git
# 编译运行
cd node_hunter
go build
./node_hunter
```
## 使用说明
1. 有`disc`、`rlpx`、`enr`三个子命令
2. `disc`子命令通过基于UDP的discover v4协议来探测以太坊网络的所有节点
3. `enr`子命令通过基于UDP探测节点的enr链接，可以获得enr链接的`seq`数据，`seq`越高暗示节点越活跃
4. `rlpx`子命令将通过基于TCP的RLPx协议与远程节点进行握手，尝试探测远程节点的操作系统、以太坊客户端版本、支持的协议类型

## 数据集
1. 探测结果保存在项目`data/storagedb`文件夹下
2. 探测数据使用**leveldb**存储
3. 利用数据库中键前缀模拟出表结构

## 数据集文件的列定义

### nodes表
> 此表存储了所有发现的节点的enode链接
1. 键格式：n<enode链接>
2. 值：发现节点的时间戳
示例：`nenode://f58fccd263ba322412ff3724466bbd774d3018b7fa00c88750b59c27e6079885fa01c97245adcbba7a1094ff8e5fda8071a283a01dab5ce72948f2cd9702ead5@195.176.181.148:30303`

### relation表
> 此表存储所有节点间的认识关系
1. 键格式：<日期>r<from节点记录><to节点记录>，代表`from`节点认识`to`节点
2. 值：发现此认识关系的时间戳
示例：`2021-12-24renode://f58fccd263ba322412ff3724466bbd774d3018b7fa00c88750b59c27e6079885fa01c97245adcbba7a1094ff8e5fda8071a283a01dab5ce72948f2cd9702ead5@195.176.181.148:30303enode://6f04d3be3ccc7fabc1e216d6f85be945e991ee9948204e2597b29c74ca334993ccf6303e9209ce52d1b73b0b7a168efb9c11284c281c75aa852b1f73895556d8@94.79.55.28:30000`

### enr表
> 此表存储所有可以查询到的enr记录
1. 键格式：<日期>e<enode链接>
2. 值：<时间戳><e或i><错误信息 或 enr链接>
示例：`2021-12-24eienr:-Je4QKlBNWb6gtKMJly2VdlSNWaSwHYhFn-mNxguWPf_OhSEfAyB-P99s2vM6UYkmqz8_0KKoDZ0gpHaHDrHB8qR5hMlg2V0aMfGhCDDJ_yAgmlkgnY0gmlwhJUHEImJc2VjcDI1NmsxoQNebyKoLEo63A7E3txx6WqWR0uCfXXA0ZDmPsBcgqAfe4N0Y3CCdl-DdWRwgnZf`
示例：`2021-12-24eeRPC timeout`

### rlpx表
> 此表存储所有可以查询到的节点的元数据
1. 键格式：<日期>x<enode链接>
2. 值：<时间戳><e或i><错误信息 或 元数据>
3. 元数据：后面两列分别为客户端版本信息，支持的各个协议
  * 整体格式：<客户端信息>空格<各个协议>
  * 客户端信息:`<客户端类型>/<版本号>`
  * 各个协议: 各个协议间以逗号分隔，`<协议名>/<协议版本号>`
4. 错误信息
  * 错误信息为`too many peers`说明此节点连接超过50个节点
  * 注：并不能直接判断连接超过50个节点，geth客户端的默认配置是超过50个节点返回此错误信息
示例：`2021-12-24xiGeth/v1.10.13-stable/linux-amd64/go1.17.5 les/2,les/3,les/4`
示例：`2021-12-24xetoo many peers`
