# **robot**
An amazing telegram robot.


### **Feature**
Basically, you can talk with the robot with both ```Chinese``` and ```English``` because it has 4 Artificial Intelligence inside:

* [Turing Robot](http://www.tuling123.com/)
* [Mitsuku Chatbot](http://www.mitsuku.com/)
* [Microsoft's Chatbot Xiaoice](http://www.msxiaoice.com/)
* [Qingyunke chatbot](http://api.qingyunke.com/)


Additionally, it supports the following commands:

```
/alarm - set an alarm
/alarms - show all of your alarms
/rmalarm - remove an alarm
/evolve	- self evolution of me
/memo - save a memo
/memos - show all of your memos
/rmmemo - remove a memo
/movie - find movie download links
/show - find American show download links
/trans - translate words between english and chinese
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
* create a telegram robt follow the [reference](https://core.telegram.org/bots)
* create your own configure file from the ```config.json``` file

After that, you have two options to run the robot:

* Execute the binary file:
```
/path/to/file/robot -c /path/to/config.json
```
* Or clone the object and build it yourself (```golang``` required):
```
go get github.com/evolsnow/robot
go build github.com/evolsnow/robot
/path/to/file/robot -c /path/to/config.json
```

### **Development**
This project is written in [go1.5](https://golang.org/doc/install).

Contact me for further questions: [@evolsnow](https://telegram.me/evolsnow)

