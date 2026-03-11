class ModalManager {
	static open(modalId) {
		const modal = document.getElementById(modalId);
		if (modal) {
			modal.classList.remove("hidden");
			document.body.style.overflow = "hidden";
		}
	}

	static close(modalId) {
		const modal = document.getElementById(modalId);
		if (modal) {
			modal.classList.add("hidden");
			document.body.style.overflow = "";
		}
	}

	static initializeControls() {
		document.querySelectorAll(".modal-close").forEach((btn) => {
			btn.addEventListener("click", (e) => {
				const modalId = btn.dataset.modal;
				this.close(modalId);
			});
		});

		document.querySelectorAll("[data-modal]").forEach((btn) => {
			btn.addEventListener("click", (e) => {
				if (
					btn.classList.contains("btn-secondary") ||
					e.target.classList.contains("modal-close")
				) {
					const modalId = btn.dataset.modal;
					if (modalId) {
						this.close(modalId);
					}
				}
			});
		});

		document.querySelectorAll(".modal").forEach((modal) => {
			modal.addEventListener("click", (e) => {
				if (e.target === modal) {
					this.close(modal.id);
				}
			});
		});

		document.addEventListener("keydown", (e) => {
			if (e.key === "Escape") {
				const openModals = document.querySelectorAll(
					".modal:not(.hidden)",
				);
				if (openModals.length > 0) {
					this.close(openModals[openModals.length - 1].id);
				}
			}
		});
	}
}

export default ModalManager;
