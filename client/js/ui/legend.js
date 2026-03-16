class LegendManager {
	static containerId = "legend";

	static statusLegend = [
		{
			id: "completed",
			name: "Выполненное ТО",
			description: "Техническое обслуживание завершено (light-green)",
			class: "status-completed",
		},
		{
			id: "current",
			name: "Текущее ТО",
			description: "Требует выполнения (просрочено ≤5 дней) (orange)",
			class: "status-current",
		},
		{
			id: "overdue",
			name: "Просроченное ТО",
			description: "Критичное (просрочено >5 дней) (red)",
			class: "status-overdue",
		},
		{
			id: "planned",
			name: "Предстоящее ТО",
			description: "Запланировано на будущее (blue)",
			class: "status-upcoming",
		},
	];

	static render() {
		const container = document.getElementById(this.containerId);
		container.innerHTML = `
            <div class="legend-title">Легенда</div>
            <div class="legend-items">
                ${this.statusLegend
					.map(
						(status) => `
                    <div class="legend-item ${status.class}">
                        <div class="legend-indicator square"></div>
                        <div class="legend-text">
                            <strong>${status.name}</strong><br>
                            <small>${status.description}</small>
                        </div>
                    </div>
                `,
					)
					.join("")}
            </div>
        `;
	}
}

export default LegendManager;
