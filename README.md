# NODE HUNTER
**以太坊网络探测工具**
## 编译运行
```
# 克隆项目
git clone https://github.com/Evolution404/node_hunter.git
# 编译运行
cd node_hunter
go build
./node_hunter
```
## 使用说明
1. 有`disc`和`rlpx`两个子命令
2. `disc`子命令通过基于UDP的discover v4协议来探测以太坊网络的所有节点
2. `rlpx`子命令将通过基于TCP的RLPx协议与远程节点进行握手，尝试探测远程节点的操作系统、以太坊客户端版本、支持的协议类型

## 数据集
1. 探测结果保存在项目`data`文件夹下
2. ndoes文件保存所有探测到的节点
3. relations文件保存各个节点的识别关系
