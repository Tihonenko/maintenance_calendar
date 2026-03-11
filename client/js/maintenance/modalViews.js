import ModalManager from "../ui/modalManager.js";
import DateUtils from "../utils/dateUtils.js";

class ModalViews {
	static showActionsModal(event, actions, isCompleted = false) {
		const titleEl = document.getElementById("actionsModalTitle");
		const listEl = document.getElementById("actionsList");

		const typeDisplay = `${event.type_code || ""} - ${event.type_name || "Unknown"}`;
		titleEl.textContent = isCompleted
			? `Выполненные действия: ${typeDisplay}`
			: `Действия для: ${typeDisplay}`;

		listEl.innerHTML = "";

		if (!actions || actions.length === 0) {
			listEl.innerHTML =
				'<p class="text-muted">Нет действий для этого ТО</p>';
		} else {
			actions.forEach((action) => {
				const itemEl = document.createElement("div");
				itemEl.className = action.is_passed
					? "action-item checked"
					: "action-item";

				let html = `
                    <div class="action-item-header">
                        ${
							isCompleted
								? `
                            <input type="checkbox" class="action-checkbox" ${action.is_passed ? "checked" : ""} disabled>
                        `
								: ""
						}
                        <div class="action-item-content">
                            <div class="action-system-node">${action.system_node || "Unknown"}</div>
                            <div class="action-description">${action.description || ""}</div>
                `;

				if (action.comment && isCompleted) {
					html += `
                        <div class="action-comment">
                            <div class="action-comment-label">Примечание:</div>
                            <div class="action-comment-text">"${action.comment}"</div>
                        </div>
                    `;
				}

				html += `
                        </div>
                    </div>
                `;

				itemEl.innerHTML = html;
				listEl.appendChild(itemEl);
			});
		}

		ModalManager.open("actionsModal");
	}

	static showPlanMaintenanceModal(event, onSave) {
		const form = document.getElementById("planMaintenanceForm");
		const dateInput = document.getElementById("plannedDate");
		const dateRangeEl = document.getElementById("dateRange");
		const saveBtn = document.getElementById("savePlanBtn");

		const calculated = DateUtils.parse(event.calculated_date);
		const minDate = DateUtils.addDays(calculated, -2);
		const maxDate = DateUtils.addDays(calculated, 2);
		const scheduled = event.scheduled_date
			? DateUtils.parse(event.scheduled_date)
			: calculated;

		dateInput.min = DateUtils.format(minDate);
		dateInput.max = DateUtils.format(maxDate);
		dateInput.value = DateUtils.format(scheduled);

		const newSaveBtn = saveBtn.cloneNode(true);
		saveBtn.parentNode.replaceChild(newSaveBtn, saveBtn);

		newSaveBtn.addEventListener("click", () => {
			const date = dateInput.value;
			if (date) {
				onSave(date);
			}
		});

		ModalManager.open("planMaintenanceModal");
	}

	static showCompleteMaintenanceModal(
		event,
		actions,
		currentMileage,
		currentHours,
		onSave,
	) {
		const form = document.getElementById("completeMaintenanceForm");
		const dateInput = document.getElementById("completionDate");
		const mileageInput = document.getElementById("completionMileage");
		const hoursInput = document.getElementById("completionHours");
		const checklistContainer =
			document.getElementById("checklistContainer");
		const confirmBtn = document.getElementById("confirmCompleteBtn");

		const today = DateUtils.today();
		dateInput.value = DateUtils.format(today);
		dateInput.max = DateUtils.format(today);
		mileageInput.value = currentMileage || "";
		hoursInput.value = currentHours || "";

		if (actions && actions.length > 0) {
			checklistContainer.innerHTML = `
                <div class="checklist-container">
                    <div class="checklist-title">✓ Чек-лист действий</div>
                    <div class="checklist-items">
                        ${actions
							.map(
								(action) => `
                            <div class="checklist-item">
                                <input type="checkbox" class="checklist-item-checkbox" data-action-id="${action.id}">
                                <div class="checklist-item-info">
                                    <div class="checklist-item-node">${action.system_node}</div>
                                    <div class="checklist-item-description">${action.description}</div>
                                </div>
                                <input type="text" class="checklist-item-input" placeholder="Комментарий (опционально)" data-comment-for="${action.id}">
                            </div>
                        `,
							)
							.join("")}
                    </div>
                </div>
            `;
		} else {
			checklistContainer.innerHTML = "";
		}

		const newConfirmBtn = confirmBtn.cloneNode(true);
		confirmBtn.parentNode.replaceChild(newConfirmBtn, confirmBtn);

		newConfirmBtn.addEventListener("click", () => {
			const date = dateInput.value;
			const mileage = parseFloat(mileageInput.value);
			const hours = parseFloat(hoursInput.value);

			if (!date || isNaN(mileage) || isNaN(hours)) {
				alert("Пожалуйста, заполните все поля");
				return;
			}

			const checklist = [];
			if (actions && actions.length > 0) {
				document
					.querySelectorAll(".checklist-item-checkbox")
					.forEach((checkbox) => {
						const actionId = parseInt(checkbox.dataset.actionId);
						const commentInput = document.querySelector(
							`[data-comment-for="${actionId}"]`,
						);
						checklist.push({
							action_id: actionId,
							is_passed: checkbox.checked,
							comment: commentInput?.value || "",
						});
					});
			}

			onSave({
				date,
				mileage,
				hours,
				checklist,
			});
		});

		ModalManager.open("completeMaintenanceModal");
	}

	static showSeasonalModal(vehicleId, onSave) {
		const form = document.getElementById("seasonalForm");
		const dateInput = document.getElementById("seasonalDate");
		const saveBtn = document.getElementById("saveSeasonalBtn");

		const today = DateUtils.today();
		dateInput.min = DateUtils.format(today);
		dateInput.value = DateUtils.format(today);

		const newSaveBtn = saveBtn.cloneNode(true);
		saveBtn.parentNode.replaceChild(newSaveBtn, saveBtn);

		newSaveBtn.addEventListener("click", () => {
			const date = dateInput.value;
			if (date) {
				onSave(date);
			}
		});

		ModalManager.open("seasonalModal");
	}

	static showDayEventsModal(date, events, getEventStatus) {
		const titleEl = document.getElementById("dayEventsModalTitle");
		const listEl = document.getElementById("dayEventsList");

		titleEl.textContent = DateUtils.formatDisplay(date);
		listEl.innerHTML = "";

		if (!events || events.length === 0) {
			listEl.innerHTML =
				'<p class="text-muted">Нет событий на этот день</p>';
		} else {
			events.forEach((event) => {
				const eventEl = document.createElement("div");
				const status = getEventStatus
					? getEventStatus(event)
					: "planned";

				eventEl.className = `calendar-event event-${status}`;
				eventEl.textContent = `${event.type_code || ""} - ${event.type_name}`;
				eventEl.title = event.type_name;
				eventEl.dataset.eventId = event.id;
				eventEl.dataset.recordId = event.id;
				eventEl.style.cursor = "pointer";
				eventEl.style.marginBottom = "5px";
				eventEl.style.padding = "8px";

				listEl.appendChild(eventEl);
			});
		}

		ModalManager.open("dayEventsModal");
	}
}

export default ModalViews;
