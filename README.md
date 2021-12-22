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
1. 探测结果保存在项目`data`文件夹下
2. ndoes文件保存所有探测到的节点
3. relations文件保存各个节点的识别关系

## 数据集文件的列定义

### nodes文件
1. 时间戳
2. 各个节点的enode链接

### relation文件
1. 时间戳
2. 节点的enode链接
3. 此节点认识的节点个数
4. 所有认识的节点，以空格分隔

### enr文件
1. 时间戳
2. 节点的enode链接
3. info或error
4. 前方若为info，则后面两列分别为节点记录序号，enr记录
5. 前方若为error，则后方为错误信息

### rlpx文件
1. 时间戳
2. 节点的enode链接
3. info或error
4. 前方若为info，则后面两列分别为客户端版本信息，支持的各个协议
  * 客户端版本:`<客户端类型>/<版本号>`
  * 各个协议:各个协议间以逗号分隔，`<协议名>/<协议版本号>`
5. 前方若为error，则后方为错误信息
  * 错误信息为`too many peers`说明此节点连接超过50个节点
  * 注：并不能直接判断连接超过50个节点，geth客户端的默认配置是超过50个节点返回此错误信息
