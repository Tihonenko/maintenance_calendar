class TruckSelector {
	constructor() {
		this.selectElement = document.getElementById("truckSelect");
		this.vehicles = [];
		this.selectedVehicleId = null;
	}

	setVehicles(vehicles) {
		this.vehicles = vehicles || [];
		this.renderOptions();
	}

	renderOptions() {
		this.selectElement.innerHTML = '<option value="">Все машины</option>';

		this.vehicles.forEach((vehicle) => {
			const option = document.createElement("option");
			option.value = vehicle.id;
			option.textContent = vehicle.vin;
			this.selectElement.appendChild(option);
		});
	}

	getSelectedVehicleId() {
		const value = this.selectElement.value;
		return value ? parseInt(value) : null;
	}

	setSelected(vehicleId) {
		if (vehicleId) {
			this.selectElement.value = vehicleId;
		} else {
			this.selectElement.value = "";
		}
		this.selectedVehicleId = vehicleId;
	}

	onChange(callback) {
		this.selectElement.addEventListener("change", () => {
			this.selectedVehicleId = this.getSelectedVehicleId();
			callback(this.selectedVehicleId);
		});
	}
}

export default TruckSelector;
