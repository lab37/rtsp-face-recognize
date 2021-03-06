# RTSPtoFaceRec
一个老掉牙的小蚁摄像机 加一台老老掉牙的破电脑  制作一个人脸识别小工具。

为什么有这个想法，因为小米智能锁，上面有摄像头，带人脸识别功能，但是，但是，但是。。。。。是要钱的。而且还不是一次买断那种。

这两天在杂物堆里找到了一个尘封多年的小蚁的摄像头，灵光一现，于是就有了这个Repo。


## 工具和原理
下面是用到的工具和简要的步骤

- 一个小蚁YHS-113/17CNY 166WR摄像头，好几年的东西了，闲鱼上很多。40块一个吧差不多。

- 用这个项目代理摄像头的rtsp流(https://github.com/aler9/rtsp-simple-server), 直接用发布的编译好的二进制程序就行。
  
- 用这个项目(https://github.com/alienatedsec/yi-hack-v5), 把小蚁摄像头刷一下固件，这样就开启了摄像头的RTSP服务。

- 参考这个项目(https://github.com/deepch/RTSPtoImage), 编写程序来连接小蚁摄像头的RTSP服务，把视频流转成一张张的照片。

- 用这个项目(https://github.com/esimov/pigo), 检测这些照片中是否有人脸。

- 如果有人脸就把这个图片交给这个项目(https://github.com/Kagami/go-face), 让它识别出人脸是谁。

## 来，跟着我，开始做

### 1.给小蚁刷上新固件
- 小蚁的型号在摄像头后面的二维码贴上，我用的是yhs-113型号(另一种说法就是二维码后面的17CNY166WR)
- 从这个地方(https://github.com/alienatedsec/yi-hack-v5), 的releases页面下载编译好的固件, 我这个型号的固件是yi_dome, 下载yi_dome_0.3.2.tgz这个文件。
下载下来后解压好。 
- 把解压出来的yi-hack-v5文件夹、home_y18、rootfs_y18文件复制到内存卡的根目录（fat32格式）。
- 摄像机不用重置，先配置好连到网络能正常使用以后，再关闭摄像机，插上内存卡后再给摄像机上电。过两三分钟系统就刷好了。
- 内存卡里面的yi-hack-v5不能删, 这个摄像机要用。home_y18和rootfs_y18可以不删(摄像机断电重开机后不会反复刷系统，因为刷了这个固件的摄像机已经被改了默认刷机文件名了), 记住不要不要不要删yi-hack-v5文件夹（这个文件夹里有很多系统要用的文件。home_y18和rootfs_y18爱删不删, 刷完就没有用了。)
- web管理地址是：http://ip:8080, 多了很功能, 上去一看就能明白。
```bash
rstp流地址：
主码流高清: http://ip:554/ch0_0.h264
辅码流 : http://ip:554/ch0_1.h264
音频流 : http://ip:554/ch0_3.h264
ssh服务的默认root密码为空

17CN就海思hi3518v200的处理器，18M的内存。15M固态。
0.3.2的固件不是很稳定，最好关掉ssh，云录像等不用的服务。只留RTSP和httpd两个。
```

### 2.安装上面那些库
#### ubunbu20.04
只提供这一个系统下的安装方法, 因为我自己就是用的这个系统, 为什么不用新出的ubuntu22.04, 不是说不能用, 是因为太新了, 怕有问题麻烦。
我电脑安装的是ubuntu20.04 server系统, 不是ubuntu桌面系统, 因为电脑很垃圾带不动桌面.
- 安装完操作系统后连上网络, 执行下面的命令完成系统基础程序的安装和配置：
```bash
sudo apt update  && sudo apt upgrade -y
sudo apt install golang  git  ffmpeg  -y
```
ubuntu20.04默认安装的golang不是最新的, 这里要求最新的golang, 为什么要用golang, 因为我只有golang用的熟。
升级golang的话要先sudo apt remove golang删除已经安装的, 然后apt autoremove  && apt autoclean
手动下载golang的安装包(网站：https://studygolang.com/dl),
```bash
wget  https://studygolang.com/dl/golang/go1.18.3.linux-amd64.tar.gz
```
然后解压安装
```bash
tar xf go1.18.3.linux-amd64.tar.gz -C /usr/local
sudo ln -sf /usr/local/go/bin/* /usr/bin/
```
然后把环境变量加到/etc/profile里
```bash
vim /etc/profile 在最后面添加下面两行
export GOPATH="$HOME/go"
export PATH=$PATH:/usr/local/go/bin
然后执行命令：
source  /etc/profile
sudo apt install gcc g++
go  env  -w GOPROXY=https://goproxy.cn,direct
mkdir  -p   ~/go/src
```
- 上面那个go-face里面用到了dlib, 用下面的命令给它安装依赖：
```bash
sudo apt install libdlib-dev libblas-dev liblapack-dev libjpeg-turbo8-dev libatlas-base-dev -y
```

- 上面那个RTSPtoImage里面用到了一些ffmpeg的库, 用下面的命令给它安装依赖：
```bash
sudo apt install libavcodec-dev  libavformat-dev  libavresample-dev  libswscale-dev  -y
```
- 下载https://github.com/aler9/rtsp-simple-server的最新二制程序, 在配置文件中找到对应位置, 加入如下配置。
```bash
paths:
  gate:
    source: rtsp://192.168.31.153:8554/gate
以后访问摄像头将使用rtsp://ubuntuIP:8554/gate
```
- 把rtsp-simple-server加入开机启动
```bash
sudo mv rtsp-simple-server /usr/local/bin/
sudo mv rtsp-simple-server.yml /usr/local/etc/

创建服务:
sudo tee /etc/systemd/system/rtsp-simple-server.service >/dev/null << EOF
[Unit]
Wants=network.target
[Service]
ExecStart=/usr/local/bin/rtsp-simple-server /usr/local/etc/rtsp-simple-server.yml
[Install]
WantedBy=multi-user.target
EOF

启动服务:
sudo systemctl enable rtsp-simple-server
sudo systemctl start rtsp-simple-server
```
行了, 搞定！

### 3.编写程序把上面的步骤粘起来
代码就在上面, 不解释了, 该说的都在代码的注释里。代码是用golang语言编写的, 因为我只会那么一点golang........

- 配置文件config.json解释一下：
```bash
{
    "imgFileName": "/home/lab37/rtspImg.jpg",
    "mqttServer": "192.168.31.153:1883",
    "mqttUserName": "lab37",
    "mqttPassword": "142857"  
}
```


- 配置文件face-data.json解释一下：
```bash
[{"name":"不认识","descriptor":[-0.086402895,0.062449184,0.025281772,-0.009759135,-0.076491865,-0.064023325,-0.040512057,-0.157585205,0.137831153,-0.106937215,0.24599106,-0.12431721,-0.191693885,-0.08147765,-0.11236507,0.255513065,-0.23072328,-0.112731145,-0.03360492,-0.032514189,0.07166216,0.014985042,0.003708921,0.088701943,-0.119844575,-0.324208085,-0.074863277,-0.118751262,-0.021223529,-0.073263458,-0.001774842,0.13434302,-0.14869849,-0.041081177,0.040243549,0.060301265,-8.71335E-05,-0.040631263,0.215542345,0.001988182,-0.22200241,0.036800291,0.104752845,0.265269695,0.129051788,0.039235579,0.005628372,-0.147551495,0.12083431,-0.14273316,-0.034898025,0.125377785,0.087508404,0.069152422,0.021131133,-0.13263075,0.060253563,0.07951126,-0.188309025,0.022543276,0.080808008,-0.115117245,0.009295596,-0.07170093,0.21518027,0.028408698,-0.109116545,-0.131378645,0.10627839,-0.14796075,-0.10411008,0.043264372,-0.110221635,-0.191448145,-0.36577347,-0.030245966,0.364821325,0.098646549,-0.190443975,0.080678125,0.023957553,-0.053486413,0.130695993,0.19222275,-0.022794363,0.019647024,-0.06424305,-0.027084,0.174070805,-0.043240131,-0.014647769,0.21065591,-0.038219288,0.032520585,0.010140643,0.017078017,-0.110517103,0.05279175,-0.085269262,-0.002775514,0.016622106,-0.02873386,0.011908661,0.061259529,-0.15876859,0.112683435,-0.030069245,0.003977443,-0.020185009,0.034421181,-0.10986702,-0.026641504,0.141332565,-0.242618605,0.195348715,0.16779973,0.034253812,0.129698445,0.117613575,0.097841295,-0.05830558,0.00868199,-0.185684965,-0.04591856,0.131664145,-0.064163616,0.116699835,-0.005825664]
},
{"name":"杨老师","descriptor":[-0.1159001,0.09263086,0.0035315836,-0.05554646,-0.12823397,-0.06515299,-0.056935456,-0.16845298,0.17980391,-0.12318635,0.24090745,-0.14977786,-0.17863044,-0.07701772,-0.16076702,0.31423423,-0.24079233,-0.10512808,-0.02752084,-0.04456017,0.04039625,-0.0074459766,0.021002386,0.13501738,-0.14683314,-0.3867429,-0.060152464,-0.12205729,-0.06866231,-0.06346952,0.002346863,0.19863264,-0.15344058,-0.026065439,0.03514251,0.08698663,0.017542275,-0.03569204,0.20831482,0.0054261843,-0.2556521,0.021006221,0.1502506,0.25756842,0.116998866,0.019827325,0.03345234,-0.13459077,0.144613,-0.14897116,-0.07909307,0.10195791,0.060774878,0.019279614,0.0580177,-0.13151145,0.042931125,0.09834204,-0.18176971,-0.026620638,0.052940886,-0.09984687,-0.026971336,-0.12301793,0.24215722,0.07947438,-0.11810785,-0.10671544,0.14092323,-0.14055648,-0.0759744,0.023722753,-0.09704798,-0.20509142,-0.39267457,-0.040415365,0.3663631,0.14050837,-0.15795466,0.09033715,0.037092745,-0.048454795,0.092696995,0.2019112,-0.046125486,0.040454436,-0.03608166,-0.025072712,0.16190164,-0.029369472,0.017310027,0.21564318,-0.047905046,0.049919203,-0.04239881,-0.0065827826,-0.16572133,0.058782343,-0.094047084,0.021720782,-0.0022751018,-0.011840454,0.020022474,0.054937173,-0.19528738,0.06903831,-0.04523987,-0.021006031,-0.041255984,0.084276654,-0.09607051,-0.034084138,0.12846239,-0.24399525,0.16876882,0.17153835,0.011270084,0.15957515,0.1532768,0.07538282,-0.042439796,-0.0006385939,-0.16411084,-0.050656885,0.16210714,-0.12667571,0.10326011,-0.013752969]
}]
```
这个是人脸特征的数据文件,每个人的人脸特征数据由128个浮点数组成。我编写了一个程序用来根据人脸的照片生成这组数(https://github.com/lab37/generate-face-128D-tools)

### 4.安装与运行
- 使用如下命令进入到GOPATH目录, 下载本程序代码：
```bash
cd ~/go/src/
git clone  https://github.com/lab37/rtsp-face-recognize.git
cd rtsp-face-recognize/
go get
```
- 编辑config.json文件, 将url处的值替换成自己摄像头的RTSP流地址。

- 使用(https://github.com/lab37/generate-face-128D-tools), 生成自己的人脸数据, 填写到face-data.json文件中。

- 在rtsp-face-recognize目录内使用如下命令开启人脸识别程序。

```bash
为了提高读写性能, 最好用内存文件系统：
mkdir /home/lab37/faceImg
sudo mount -t tmpfs -o size=100M tmpfs /home/lab37/faceImg
系统重启后内存挂载的文件系统会消失，可以写入fstab长期挂载
在/etc/fstab文件中增加挂载配置，可以实现系统启动时自动挂载。具体如下：
sudo gedit /etc/fstab
在文件中增加如下内容并保存。
tmpfs	/home/lab37/faceImg	tmpfs	defaults,size=100M	0 0


go run *.go
```
搞定!



## 扩展功能
识别人脸后将人脸数据以MQTT协议的方式传递给home-assistant, 用home-assistant调用小爱智能音箱的TTS功能，播报人名。可以做为一个人脸识别门禁使用。
此功能已增加到代码中
1. 在ubuntu中安装MQTT服务器
```bash
sudo  apt install mosquitto
配置文件在：/etc/mosquitto/mosquitto.conf 以及/etc/mosquitto/conf.d/中以.conf结尾的文件里
添加配置文件
在/etc/mosquitto/conf.d目录下，添加配置文件myconfig.conf 配置文件：
sudo vi /etc/mosquitto/conf.d/myconfig.conf
粘入下面这些配置
#-------------------------------------------
#添加监听端口（很重要，否则只能本机访问）
listener 1883

# 关闭匿名访问，客户端必须使用用户名
allow_anonymous false

#指定 用户名-密码 文件
password_file /etc/mosquitto/pwfile.txt
#-------------------------------------------
添加账户及密码
sudo mosquitto_passwd -c /etc/mosquitto/pwfile.txt lab37
回车后连续输入2次用户密码即可
启动mosquitto
sudo service mosquitto stop
sudo service mosquitto start
```
2. 使用(https://github.com/yosssi/gmq), 这是个纯golang写的mqtt客户端库编写mqtt代码。
代码已在上面mqtt.go文件中

3. home-assistant安装MQTT集成，配置其连接到ubuntu中的mqtt服务器。
4. 安装(https://github.com/deepch/RTSPtoWebRTC), 将rtsp-simple-server的rtsp流转为webRtc提
   供给home-assistant用。
5. docker安装home-assistant
```bash
docker要先登陆: sudo  docker login
sudo docker search home-assistant
sudo docker pull homeassistant/home-assistant
sudo docker run -d --name="hass" -v /home/lab37/hass:/config -p 8123:8123 homeassistant/home-assistant
d：表示在后台运行
name：给容器设置别名（不然会随机生成，为了方便管理）
v：配置数据卷（容器内的数据直接映射到本地主机环境，参考路径配置， 意思就是把宿主机的hass目录, 映射到主机的config目录
p：映射端口（容器内的端口直接映射到本地主机端口最后便是刚才下载的镜像了，运行该容器。
查看状态
docker ps
## 启动
docker start hass
## 停止
docker stop hass
## 列出所有镜像
docker images
## 列出创建的所有容器
docker ps -a
## 删除这个容器
docker rm 5e436ae313
## 删除镜像
docker rmi name:tag| ID
```


## Tips
### 一些用来测试的命令（备忘）
```bash
在windows下用ffmpeg播放RTSP视频流
ffplay   "rtsp://192.168.31.225:554/ch0_1.h264"
 
每隔1秒截取一张图片并都按一定的规则命名来生成图片
ffmpeg -i "rtsp://192.168.31.225:554/ch0_1.h264" -y -f image2 -r 1/1 /home/lab37/faceImg/img%03d.jpg

每隔1秒截取一张指定分辨率的图片并覆盖在同一张图片上
ffmpeg -i "rtsp://192.168.31.225:554/ch0_1.h264" -y -f image2 -r 1/1  -update  1 -s 640x480 /home/lab37/faceImg/rtsp.jpg

每隔1秒截取5张指定分辨率的图片并覆盖在同一张图片上
ffmpeg -i "rtsp://192.168.31.225:554/ch0_1.h264" -y -f image2 -r 5/1 -update 1  /home/lab37/faceImg/rtsp.jpg

每隔1秒截取5张指定分辨率960*540的图片, 转成灰度图并覆盖在同一张图片上(-vf format=gray是转为灰度图)
ffmpeg -i "rtsp://192.168.31.225:554/ch0_1.h264" -y -f image2 -r 5/1 -update 1   -s 960x540  -vf format=gray  /home/lab37/faceImg/rtsp.jpg


为了提高读写性能, 最好用内存文件系统：
mkdir /home/lab37/faceImg
sudo mount -t tmpfs -o size=100M tmpfs /home/lab37/faceImg

系统重启后内存挂载的文件系统会消失，可以写入fstab长期挂载
在/etc/fstab文件中增加挂载配置，可以实现系统启动时自动挂载。具体如下：
sudo gedit /etc/fstab
在文件中增加如下内容并保存。
tmpfs	/home/lab37/faceImg	tmpfs	defaults,size=100M	0 0

只输出错误到文件
nohup command -c -b -d aaa.txt  > /dev/null 2 > log &

ffmpeg有时会异常退出, 需要监控ffmpeg运行, 编写脚本：ffmpeg2jpg.sh
timeout 20 ffmpeg -i "rtsp://127.0.0.1:8554/gate" -y -f image2 -r 3/1 -update 1   -vf format=gray  /home/lab37/faceImg/rtsp.jpg 2> /dev/null &

再编写一个监控ffmpeg的脚本, check_ff_mp_eg_live.sh
#!/bin/sh 
num=`ps -ef | grep ffmpeg | grep -v grep | wc -l`
if [ $num -lt 1 ]
then
 . /home/lab37/ffmpeg2jpg.sh
fi
上面那个.不要落了，这是一个脚本调用另一个脚本的方法，或者用source. 因为脚本名字中有ffmpeg，所以要分开写,不然麻烦, 
把脚本添加crontab
crontab -e 
*/1 * * * *  /home/lab37/check_ff_mp_eg_live.sh

Ubuntu默认没有开启cron定时任务的执行日志，需手动打开
编辑 rsyslog 配置文件，如果没有就新建一个
sudo vim /etc/rsyslog.d/50-default.conf
取消 cron 注释，变成如下（如果没有此行配置就下入如下配置）
cron.*          /var/log/cron.log
重启 rsyslog 服务
sudo service rsyslog restart
然后执行crontab的任务，比如设置一个每分钟执行一次的，
过一分钟之后就可以看到生成了 /var/log/cron.log 文件
查看没有问题后最好关掉这个日志。

hass中的xiaomi iot auto集成刷新状态很慢, 频率很低。无法满足实时性要求。只能监听绿米网关的组播信息来触发一些实时性场景。
```

