import DateUtils from "../utils/dateUtils.js";

class MaintenanceService {
	static getStatus(event) {
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

	static formatTypeName(event) {
		if (event.type_code && event.type_name) {
			return `${event.type_code} - ${event.type_name}`;
		}
		return event.type_name || "Unknown";
	}

	static canReschedule(event) {
		const isCompleted = event.ui_status === "light-green";
		return !isCompleted && !event.is_seasonal;
	}

	static canComplete(event) {
		const isCompleted = event.ui_status === "light-green";
		return !isCompleted;
	}

	static isSeasonal(event) {
		return event.is_seasonal === true;
	}

	static getAllowedReschedule(event) {
		const calculated = DateUtils.parse(event.calculated_date);
		const minDate = DateUtils.addDays(calculated, -2);
		const maxDate = DateUtils.addDays(calculated, 2);

		return {
			minDate: DateUtils.format(minDate),
			maxDate: DateUtils.format(maxDate),
			calculated: DateUtils.format(calculated),
		};
	}

	static getDaysOverdue(event) {
		return DateUtils.getDaysOverdue(
			event.scheduled_date || event.calculated_date,
		);
	}
}

export default MaintenanceService;
