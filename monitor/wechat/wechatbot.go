package wechat

//
//
//
//import (
//"github.com/liangjfblue/wxBot4g/models"
//"github.com/liangjfblue/wxBot4g/pkg/define"
//"github.com/liangjfblue/wxBot4g/wcbot"
//
//"github.com/sirupsen/logrus"
//)
//
//func HandleMsg(msg models.RealRecvMsg) {
//	logrus.Debug("MsgType: ", msg.MsgType, " ", " MsgTypeId: ", msg.MsgTypeId)
//	logrus.Info(
//		"消息类型:", define.MsgIdString(msg.MsgType), " ",
//		"数据类型:", define.MsgTypeIdString(msg.MsgTypeId), " ",
//		"发送人:", msg.SendMsgUSer.Name, " ",
//		"内容:", msg.Content.Data)
//}
//
//func main() {
//	bot := wcbot.New(HandleMsg)
//	bot.Debug = true
//	bot.Run()
//}
//
