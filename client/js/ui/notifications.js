class NotificationManager {
	static containerId = "notifications";
	static notificationDuration = 5000;

	static show(message, type = "info", duration = this.notificationDuration) {
		const container = document.getElementById(this.containerId);
		const notification = document.createElement("div");
		notification.className = `notification ${type}`;

		notification.innerHTML = `
            <div class="notification-content">${message}</div>
            <button class="notification-close">✕</button>
        `;

		const closeBtn = notification.querySelector(".notification-close");
		closeBtn.addEventListener("click", () => this.remove(notification));

		container.appendChild(notification);

		if (duration > 0) {
			setTimeout(() => this.remove(notification), duration);
		}

		return notification;
	}

	static remove(notification) {
		notification.classList.add("removing");
		setTimeout(() => {
			notification.remove();
		}, 300);
	}

	static success(message) {
		return this.show(message, "success");
	}

	static error(message) {
		return this.show(message, "error", 8000);
	}

	static warning(message) {
		return this.show(message, "warning", 6000);
	}

	static info(message) {
		return this.show(message, "info");
	}

	static clearAll() {
		const container = document.getElementById(this.containerId);
		const notifications = container.querySelectorAll(".notification");
		notifications.forEach((n) => this.remove(n));
	}
}

export default NotificationManager;
