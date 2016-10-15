# **robot**
An amazing telegram robot.

[JianShu Blog](http://www.jianshu.com/p/46c7d36a95d2)

[![Build Status](https://api.travis-ci.org/evolsnow/robot.svg?branch=master)](https://travis-ci.org/evolsnow/robot)

### **Feature**
Basically, you can talk with the robot with both ```Chinese``` and ```English``` because it has 4 Artificial Intelligence inside:

* [Turing Robot](http://www.tuling123.com/)
* [Mitsuku Chatbot](http://www.pandorabots.com/)
* [Microsoft's Chatbot Xiaoice](http://www.msxiaoice.com/)
* [Qingyunke chatbot](http://api.qingyunke.com/)


Additionally, it supports the following commands:

```
/alarm - set an alarm
/alarms - show all of your alarms
/rmalarm - remove an alarm
/memo - save a memo
/memos - show all of your memos
/rmmemo - remove a memo
/movie - find movie download links
/show - find American show download links
/trans - translate words between english and chinese
/exit - exit any interactive mode
/help - show this help message
```

Is that all?
More skills are developing and any suggestions will be greatly appreciated.

For more advanced usage, contact [@EvolsnowBot](https://telegram.me/EvolsnowBot) right away.

### **Demo**
Remember that the website can just show a part of the robot's skills.
To be precise, it's just able to talk with you...

 https://samaritan.tech
### **Installation**
You can simply try the robot in telegram as mentioned above.
But if you want to build your own robot:

* [redis](http://redis.io/download) is required.
* create a telegram robot follow the [reference](https://core.telegram.org/bots)
* create your own configure file from the ```config.json``` file

After that, you have two options to run the robot:

* Execute the binary file from [release](https://github.com/evolsnow/robot/releases):
```
/path/to/file/robot -c /path/to/config.json
```
* Or clone the object and build it yourself (```golang``` required):
```
go get -u github.com/evolsnow/robot
go build github.com/evolsnow/robot
/path/to/file/robot -c /path/to/config.json
```

You may want to add ```-d``` flag to debug.

### **Development**
This project is written in [go](https://golang.org/doc/install).

Contact me for further questions: [@evolsnow](https://telegram.me/evolsnow)

### **License**
[The MIT License (MIT)](https://raw.githubusercontent.com/evolsnow/robot/master/LICENSE)
