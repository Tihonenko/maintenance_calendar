import DateUtils from "../utils/dateUtils.js";

class CalendarManager {
	constructor() {
		const today = DateUtils.getCurrentMonth();
		this.currentMonth = today.month;
		this.currentYear = today.year;
		this.events = [];
		this.vehicles = [];
		this.selectedVehicleId = null;
	}

	setMonth(month, year) {
		this.currentMonth = month;
		this.currentYear = year;
	}

	getMonth() {
		return { month: this.currentMonth, year: this.currentYear };
	}

	nextMonth() {
		if (this.currentMonth === 12) {
			this.currentMonth = 1;
			this.currentYear++;
		} else {
			this.currentMonth++;
		}
	}

	previousMonth() {
		if (this.currentMonth === 1) {
			this.currentMonth = 12;
			this.currentYear--;
		} else {
			this.currentMonth--;
		}
	}

	goToToday() {
		const today = DateUtils.getCurrentMonth();
		this.currentMonth = today.month;
		this.currentYear = today.year;
	}

	setEvents(events) {
		this.events = events || [];
	}

	setVehicles(vehicles) {
		this.vehicles = vehicles || [];
	}

	getEventsForDate(date) {
		if (typeof date === "string") {
			date = new Date(date);
		}

		const dateString = DateUtils.format(date);
		return this.events.filter((event) => {
			const eventDate = event.scheduled_date || event.calculated_date;
			return DateUtils.format(eventDate) === dateString;
		});
	}

	getEventStatus(event) {
		if (event.ui_status) {
			const statusMap = {
				"light-green": "completed",
				green: "planned",
				orange: "current",
				red: "overdue",
			};
			return statusMap[event.ui_status] || "planned";
		}

		return "planned";
	}

	setOnMoreClick(callback) {
		this.onMoreClick = callback;
	}

	render() {
		const container = document.getElementById("calendar");
		const monthYearEl = document.getElementById("currentMonthYear");

		monthYearEl.textContent = DateUtils.getMonthYearString(
			this.currentMonth,
			this.currentYear,
		);

		container.innerHTML = "";

		const dayHeaders = document.createElement("div");
		dayHeaders.className = "calendar-day-header";
		const dayNames = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"];
		dayHeaders.innerHTML = dayNames
			.map((day) => `<div class="calendar-day-name">${day}</div>`)
			.join("");
		container.appendChild(dayHeaders);
		const days = DateUtils.getMonthDays(
			this.currentMonth,
			this.currentYear,
		);

		days.forEach((date) => {
			const dayEl = document.createElement("div");
			const isOtherMonth = date.getMonth() + 1 !== this.currentMonth;
			const isToday = DateUtils.isToday(date);

			dayEl.className = "calendar-day";
			if (isOtherMonth) dayEl.classList.add("other-month");
			if (isToday) dayEl.classList.add("today");

			const dayNumber = document.createElement("div");
			dayNumber.className = "day-number";
			dayNumber.textContent = date.getDate();
			dayEl.appendChild(dayNumber);

			const dateEvents = this.getEventsForDate(date);
			const MAX_EVENTS = 4;

			dateEvents.slice(0, MAX_EVENTS).forEach((event) => {
				const eventEl = document.createElement("div");
				eventEl.className = `calendar-event event-${this.getEventStatus(event)}`;
				eventEl.textContent = event.type_code || event.type_name;
				
				const vehicle = this.vehicles.find(v => v.id === event.vehicle_id);
				const vinInfo = vehicle && vehicle.vin ? ` (VIN: ${vehicle.vin})` : "";
				eventEl.title = `${event.type_name}${vinInfo}`;
				
				eventEl.dataset.eventId = event.id;
				eventEl.dataset.recordId = event.id;
				eventEl.style.cursor = "pointer";

				dayEl.appendChild(eventEl);
			});

			if (dateEvents.length > MAX_EVENTS) {
				const moreCount = dateEvents.length - MAX_EVENTS;
				const moreEl = document.createElement("div");
				moreEl.className = "calendar-more-events";
				moreEl.textContent = `Еще +${moreCount}`;
				moreEl.addEventListener("click", (e) => {
					e.stopPropagation();
					if (this.onMoreClick) {
						this.onMoreClick(date, dateEvents);
					}
				});
				dayEl.appendChild(moreEl);
			}

			container.appendChild(dayEl);
		});
	}
}

export default CalendarManager;
