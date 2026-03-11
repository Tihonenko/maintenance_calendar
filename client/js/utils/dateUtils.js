class DateUtils {
	static today() {
		const date = new Date();
		return new Date(date.getFullYear(), date.getMonth(), date.getDate());
	}

	static format(date) {
		if (typeof date === "string") {
			date = new Date(date);
		}
		const year = date.getFullYear();
		const month = String(date.getMonth() + 1).padStart(2, "0");
		const day = String(date.getDate()).padStart(2, "0");
		return `${year}-${month}-${day}`;
	}

	static formatDisplay(date) {
		if (typeof date === "string") {
			date = new Date(date);
		}
		const months = [
			"января",
			"февраля",
			"марта",
			"апреля",
			"мая",
			"июня",
			"июля",
			"августа",
			"сентября",
			"октября",
			"ноября",
			"декабря",
		];
		return `${date.getDate()} ${months[date.getMonth()]} ${date.getFullYear()}`;
	}

	static addDays(date, days) {
		if (typeof date === "string") {
			date = new Date(date);
		}
		const result = new Date(date);
		result.setDate(result.getDate() + days);
		return result;
	}

	static daysBetween(date1, date2) {
		if (typeof date1 === "string") date1 = new Date(date1);
		if (typeof date2 === "string") date2 = new Date(date2);

		const d1 = new Date(
			date1.getFullYear(),
			date1.getMonth(),
			date1.getDate(),
		);
		const d2 = new Date(
			date2.getFullYear(),
			date2.getMonth(),
			date2.getDate(),
		);

		const diffTime = Math.abs(d2 - d1);
		return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
	}

	static getDaysOverdue(scheduledDate) {
		if (typeof scheduledDate === "string") {
			scheduledDate = new Date(scheduledDate);
		}
		const today = this.today();
		const scheduled = new Date(
			scheduledDate.getFullYear(),
			scheduledDate.getMonth(),
			scheduledDate.getDate(),
		);

		const diffTime = today - scheduled;
		return Math.floor(diffTime / (1000 * 60 * 60 * 24));
	}

	static isToday(date) {
		if (typeof date === "string") {
			date = new Date(date);
		}
		const today = this.today();
		return date.toDateString() === today.toDateString();
	}

	static isPast(date) {
		if (typeof date === "string") {
			date = new Date(date);
		}
		const today = this.today();
		const checkDate = new Date(
			date.getFullYear(),
			date.getMonth(),
			date.getDate(),
		);
		return checkDate < today;
	}

	static isFuture(date) {
		if (typeof date === "string") {
			date = new Date(date);
		}
		const today = this.today();
		const checkDate = new Date(
			date.getFullYear(),
			date.getMonth(),
			date.getDate(),
		);
		return checkDate > today;
	}

	static getFirstDayOfMonth(month, year) {
		return new Date(year, month - 1, 1);
	}

	static getLastDayOfMonth(month, year) {
		return new Date(year, month, 0);
	}

	static getMonthDays(month, year) {
		const firstDay = this.getFirstDayOfMonth(month, year);
		const lastDay = this.getLastDayOfMonth(month, year);

		const startDate = new Date(firstDay);
		startDate.setDate(startDate.getDate() - firstDay.getDay());

		const endDate = new Date(lastDay);
		endDate.setDate(endDate.getDate() + (6 - lastDay.getDay()));

		const days = [];
		const currentDate = new Date(startDate);

		while (currentDate <= endDate) {
			days.push(new Date(currentDate));
			currentDate.setDate(currentDate.getDate() + 1);
		}

		return days;
	}

	static getMonthYearString(month, year) {
		const months = [
			"Январь",
			"Февраль",
			"Март",
			"Апрель",
			"Май",
			"Июнь",
			"Июль",
			"Август",
			"Сентябрь",
			"Октябрь",
			"Ноябрь",
			"Декабрь",
		];
		return `${months[month - 1]} ${year}`;
	}

	static parse(dateString) {
		if (typeof dateString !== "string") {
			return dateString;
		}
		return new Date(dateString + "T00:00:00");
	}

	static getCurrentMonth() {
		const today = new Date();
		return {
			month: today.getMonth() + 1,
			year: today.getFullYear(),
		};
	}

	static toISODateTime(dateInput) {
		let date;

		if (typeof dateInput === "string") {
			date = new Date(dateInput + "T00:00:00");
		} else if (dateInput instanceof Date) {
			date = dateInput;
		} else {
			return null;
		}

		return date.toISOString();
	}
}

export default DateUtils;
