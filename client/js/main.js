import ApiClient from "./api.js";

import CalendarManager from "./calendar/calendarManager.js";

import TruckSelector from "./truck/truckSelector.js";

import MaintenanceService from "./maintenance/maintenanceService.js";

import ModalViews from "./maintenance/modalViews.js";

import ModalManager from "./ui/modalManager.js";

import LegendManager from "./ui/legend.js";

import NotificationManager from "./ui/notifications.js";

import DateUtils from "./utils/dateUtils.js";



class App {

	constructor() {

		this.calendar = new CalendarManager();

		this.truckSelector = new TruckSelector();

		this.vehicles = [];

		this.currentVehicleId = null;

		this.currentModalDate = null;

		this.maintenanceTypes = [];

		this.loadingSpinner = document.getElementById("loadingSpinner");

		this.clickTimeout = null;

	}



	async init() {

		try {

			this.showLoading(true);



			ModalManager.initializeControls();

			LegendManager.render();



			this.calendar.setOnMoreClick((date, events) => {

				this.currentModalDate = date;

				ModalViews.showDayEventsModal(date, events, (event) =>

					this.calendar.getEventStatus(event),

				);

			});



			await this.loadVehicles();

			await this.loadCalendar();



			this.setupEventListeners();



			NotificationManager.success("Приложение загружено успешно");

		} catch (error) {

			console.error("Initialization error:", error);

			NotificationManager.error("Ошибка при загрузке приложения");

		} finally {

			this.showLoading(false);

		}

	}



	async loadVehicles() {

		try {

			const response = await ApiClient.getVehicles();

			this.vehicles = response.vehicles || [];

			this.truckSelector.setVehicles(this.vehicles);

		} catch (error) {

			console.error("Failed to load vehicles:", error);

			NotificationManager.error("Не удалось загрузить список машин");

		}

	}



	async loadCalendar() {

		try {

			this.showLoading(true);

			const { month, year } = this.calendar.getMonth();

			const vehicleId = this.truckSelector.getSelectedVehicleId();



			const response = await ApiClient.getCalendar(

				vehicleId,

				month,

				year,

			);

			this.calendar.setEvents(response.events || []);

			this.calendar.setVehicles(this.vehicles);

			this.calendar.render();



			const dayEventsModal = document.getElementById("dayEventsModal");

			if (

				dayEventsModal &&

				!dayEventsModal.classList.contains("hidden") &&

				this.currentModalDate

			) {

				const events = this.calendar.getEventsForDate(

					this.currentModalDate,

				);

				ModalViews.showDayEventsModal(

					this.currentModalDate,

					events,

					(event) => this.calendar.getEventStatus(event),

				);

			}

		} catch (error) {

			console.error("Failed to load calendar:", error);

			NotificationManager.error("Не удалось загрузить календарь");

		} finally {

			this.showLoading(false);

		}

	}



	setupEventListeners() {

		document.getElementById("prevMonth").addEventListener("click", () => {

			this.calendar.previousMonth();

			this.loadCalendar();

		});



		document.getElementById("nextMonth").addEventListener("click", () => {

			this.calendar.nextMonth();

			this.loadCalendar();

		});



		document.getElementById("todayBtn").addEventListener("click", () => {

			this.calendar.goToToday();

			this.loadCalendar();

		});



		const planSeasonalBtn = document.getElementById("planSeasonalBtn");

		if (planSeasonalBtn) {

			planSeasonalBtn.addEventListener("click", () => {

				this.handleCreateSeasonal();

			});

		}



		document

			.getElementById("monthInput")

			.addEventListener("change", (e) => {

				const [year, month] = e.target.value.split("-");

				if (year && month) {

					this.calendar.setMonth(parseInt(month), parseInt(year));

					this.loadCalendar();

				}

			});



		this.truckSelector.onChange(() => {

			this.calendar.goToToday();

			this.loadCalendar();

		});



		document.addEventListener("click", (e) => {

			const eventEl = e.target.closest(".calendar-event");

			if (eventEl) {

				this.handleEventClick(eventEl);

			}

		});



		document.addEventListener("dblclick", (e) => {

			const eventEl = e.target.closest(".calendar-event");

			if (eventEl) {

				this.handleEventDoubleClick(eventEl);

			}

		});

	}



	async handleEventClick(eventEl) {

		const recordId = parseInt(eventEl.dataset.recordId);

		const event = this.findEventById(recordId);



		if (!event) return;



		if (this.clickTimeout) {

			clearTimeout(this.clickTimeout);

		}



		this.clickTimeout = setTimeout(async () => {

			const isCompleted = event.ui_status === "light-green";

			if (isCompleted) {

				await this.showEventActions(recordId, true);

			} else {

				this.showEventOptions(event);

			}

			this.clickTimeout = null;

		}, 250);

	}



	async handleEventDoubleClick(eventEl) {

		if (this.clickTimeout) {

			clearTimeout(this.clickTimeout);

			this.clickTimeout = null;

		}



		const recordId = parseInt(eventEl.dataset.recordId);

		const event = this.findEventById(recordId);



		if (!event) return;



		const isCompleted = event.ui_status === "light-green";

		if (isCompleted) {

			await this.showEventActions(recordId, true);

		} else {

			await this.showEventActions(recordId, false);

		}

	}



	showEventOptions(event) {

		const options = [];



		if (

			!MaintenanceService.isSeasonal(event) &&

			MaintenanceService.canReschedule(event)

		) {

			options.push({

				label: "Перепланировать",

				action: () => this.handleReschedule(event),

			});

		}



		if (MaintenanceService.canComplete(event)) {

			options.push({

				label: "Отметить выполненным",

				action: () => this.handleComplete(event),

			});

		}



		if (MaintenanceService.isSeasonal(event)) {

			options.push({

				label: "Перепланировать сезонное",

				action: () => this.handleSeasonalReschedule(event),

			});

		}



		if (options.length > 0) {

			const choice =

				options.length === 1

					? 0

					: parseInt(

							prompt(

								`Выберите действие:\n${options.map((o, i) => `${i + 1}. ${o.label}`).join("\n")}`,

							),

						) - 1;



			if (!isNaN(choice) && options[choice]) {

				options[choice].action();

			}

		}

	}



	async showEventActions(recordId, isCompleted) {

		try {

			this.showLoading(true);

			

			let actions = [];

			

			if (isCompleted) {

				const response = await ApiClient.getMaintenanceDetails(recordId);

				actions = response.items || [];

			} else {

				const response = await ApiClient.getEventWithActions(recordId);

				actions = response.event || [];

			}



			const event = this.findEventById(recordId);

			if (event) {

				const vehicle = this.vehicles.find(

					(v) => v.id === event.vehicle_id,

				);

				ModalViews.showActionsModal(event, actions, isCompleted, vehicle);

			}

		} catch (error) {

			console.error("Failed to load actions:", error);

			NotificationManager.error("Не удалось загрузить действия");

		} finally {

			this.showLoading(false);

		}

	}



	handleReschedule(event) {

		ModalViews.showPlanMaintenanceModal(event, async (newDate) => {

			try {

				this.showLoading(true);

				await ApiClient.rescheduleMaintenance(event.id, {

					new_scheduled_date: DateUtils.toISODateTime(newDate),

				});

				NotificationManager.success("ТО перепланировано успешно");

				ModalManager.close("planMaintenanceModal");

				await this.loadCalendar();

			} catch (error) {

				console.error("Reschedule error:", error);

				NotificationManager.error(`Ошибка: ${error.message}`);

			} finally {

				this.showLoading(false);

			}

		});

	}



	async handleComplete(event) {

		try {

			const vehicle = this.vehicles.find(

				(v) => v.id === event.vehicle_id,

			);

			const currentMileage = vehicle?.total_mileage || 0;

			const currentHours = vehicle?.total_engine_hours || 0;



			const response = await ApiClient.getEventWithActions(event.id);

			const actions = response.event || [];



			ModalViews.showCompleteMaintenanceModal(

				event,

				actions,

				currentMileage,

				currentHours,

				async (data) => {

					try {

						this.showLoading(true);

						const completeData = {

							completion_date: DateUtils.toISODateTime(data.date),

							mileage: data.mileage,

							engine_hours: data.hours,

							checklist: data.checklist,

						};



						await ApiClient.completeMaintenance(

							event.id,

							completeData,

						);

						NotificationManager.success(

							"ТО отмечено как выполненное",

						);

						ModalManager.close("completeMaintenanceModal");

						await this.loadCalendar();

					} catch (error) {

						console.error("Complete error:", error);

						NotificationManager.error(`Ошибка: ${error.message}`);

					} finally {

						this.showLoading(false);

					}

				},

			);

		} catch (error) {

			console.error("Failed to prepare completion:", error);

			NotificationManager.error(

				"Не удалось подготовить форму выполнения",

			);

		}

	}



	handleSeasonalReschedule(event) {

		ModalViews.showSeasonalModal(event.vehicle_id, async (newDate) => {

			try {

				this.showLoading(true);

				await ApiClient.createSeasonal({

					vehicle_id: event.vehicle_id,

					scheduled_date: DateUtils.toISODateTime(newDate),

				});

				NotificationManager.success("Сезонное ТО перепланировано");

				ModalManager.close("seasonalModal");

				await this.loadCalendar();

			} catch (error) {

				console.error("Seasonal reschedule error:", error);

				NotificationManager.error(`Ошибка: ${error.message}`);

			} finally {

				this.showLoading(false);

			}

		});

	}



	handleCreateSeasonal() {

		const vehicleId = this.truckSelector.getSelectedVehicleId();

		if (!vehicleId) {

			NotificationManager.error("Выберите самосвал для планирования сезонного ТО");

			return;

		}



		ModalViews.showSeasonalModal(vehicleId, async (newDate) => {

			try {

				this.showLoading(true);

				await ApiClient.createSeasonal({

					vehicle_id: parseInt(vehicleId),

					scheduled_date: DateUtils.toISODateTime(newDate),

				});

				NotificationManager.success("Сезонное ТО запланировано");

				ModalManager.close("seasonalModal");

				await this.loadCalendar();

			} catch (error) {

				console.error("Create seasonal error:", error);

				NotificationManager.error(`Ошибка: ${error.message}`);

			} finally {

				this.showLoading(false);

			}

		});

	}



	findEventById(recordId) {

		return this.calendar.events.find((e) => e.id === recordId);

	}



	showLoading(show) {

		if (show) {

			this.loadingSpinner.classList.remove("hidden");

		} else {

			this.loadingSpinner.classList.add("hidden");

		}

	}

}



document.addEventListener("DOMContentLoaded", async () => {

	const app = new App();

	await app.init();



	window.app = app;

});

