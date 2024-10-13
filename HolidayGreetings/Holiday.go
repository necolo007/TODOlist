package HolidayGreetings

import (
	"context"
	"fmt"
	"github.com/Lofanmi/chinese-calendar-golang/calendar"
	staffv1 "github.com/hduhelp/api_open_sdk/campusapis/staff/v1"
	"github.com/hduhelp/api_open_sdk/transfer"
	"github.com/hduhelp/wechat_mp_server/hub"
	"github.com/hduhelp/wechat_mp_server/module/templateMessage"
	"github.com/hduhelp/wechat_mp_server/utils"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log"
	"time"
)

func (m *module) sendNationalDayWishes() {
	today := time.Now()
	month := int(today.Month())
	day := today.Day()
	if month == 10 && day == 1 {
		List := m.GenerateList(context.Background())
		templateMessageModule, _ := hub.GetModule(hub.NewModuleID("atom", "templateMessage"))
		templateMessageSender := templateMessageModule.Instance.(*templateMessage.Module)
		wishesCount := len(m.wishes.NDWishes)
		const MaxWishLen = 170
		for i, listStu := range List {
			unionID := hub.ConvertHDUIdToWechatUnionId(context.Background(), listStu.StaffId)
			if unionID != "" {
				fullMsg := fmt.Sprintf("\n%s同学！祝你国庆快乐呀！\n"+m.wishes.NDWishes[i%wishesCount], listStu.StaffName)
				msgList := utils.AddTip(utils.CutMsg(fullMsg, MaxWishLen*3, 0), "接上条\n")
				for _, msg := range msgList {
					fmt.Println(msg)
					templateMessageSender.PushMessage(&templateMessage.TemplateMessage{
						Message: &message.TemplateMessage{
							ToUser: unionID,
							Data: map[string]*message.TemplateDataItem{
								"keyword1": {
									Value: "叮咚~这是一条国庆祝福!\n\n",
								},
								"keyword2": {
									Value: msg,
								},
							},
							TemplateID: "_HtuD7TrFKxquwizJwICXv4sWg5AeZBvaHBIRvYKeKk",
						},
					})
				}
			}
		}
	}
}

func (m *module) GenerateList(ctx context.Context) []*staffv1.PersonInfo {
	var infos []*staffv1.PersonInfo
	err := transfer.Get(ctx, "salmon_base", "/student/Holidays", nil, utils.Someone).
		EndStruct(&infos)
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	return infos
}

// DateConvert 农历转洋历
func DateConvert(year, month, day int64) (int64, int64, int64) {
	lunar := calendar.ByLunar(year, month, day, 0, 0, 1, JudgeLeapYear(year))
	solar := lunar.Solar
	return solar.GetYear(), solar.GetMonth(), solar.GetDay()
}

// JudgeLeapYear 判断是否是闰年
func JudgeLeapYear(year int64) bool {
	if (year%4 == 0 && year%100 != 0) || (year%400 == 0) {
		return true
	} else {
		return false
	}
}
