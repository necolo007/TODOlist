package HolidayGreetings

import "time"

func (m *module) sendHolidayWishes() {
	today := time.Now()
	month := int(today.Month())
	day := today.Day()
	if month == 10 && day == 1 {

	}
}
