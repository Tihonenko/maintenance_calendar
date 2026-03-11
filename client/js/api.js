const API_BASE_URL = "http://localhost:3030/api/v1";

class ApiClient {
	static async getVehicles() {
		return this._request("GET", "/vehicles");
	}

	static async getVehicle(id) {
		return this._request("GET", `/vehicles/${id}`);
	}

	static async getCalendar(vehicleId = null, month, year) {
		let url = `/calendar?month=${month}&year=${year}`;
		if (vehicleId) {
			url += `&vehicle_id=${vehicleId}`;
		}
		return this._request("GET", url);
	}

	static async getMaintenanceTypes() {
		return this._request("GET", "/maintenance/types");
	}

	static async getMergedActions(typeId) {
		return this._request("GET", `/maintenance/types/${typeId}/actions`);
	}

	static async getEventWithActions(recordId) {
		return this._request(
			"GET",
			`/maintenance/${recordId}/event-with-actions`,
		);
	}

	static async getMaintenanceDetails(recordId) {
		return this._request("GET", `/maintenance/${recordId}/details`);
	}

	static async completeMaintenance(recordId, data) {
		return this._request("POST", `/maintenance/${recordId}/complete`, data);
	}

	static async rescheduleMaintenance(recordId, data) {
		return this._request(
			"PUT",
			`/maintenance/${recordId}/reschedule`,
			data,
		);
	}

	static async createSeasonal(data) {
		return this._request("POST", "/maintenance/seasonal", data);
	}

	static async _request(method, endpoint, data = null) {
		const options = {
			method,
			headers: {
				"Content-Type": "application/json",
			},
		};

		if (data) {
			options.body = JSON.stringify(data);
		}

		try {
			const response = await fetch(`${API_BASE_URL}${endpoint}`, options);

			if (!response.ok) {
				const error = await response.json().catch(() => ({}));
				throw new Error(
					error.error || `HTTP Error: ${response.status}`,
				);
			}

			return await response.json();
		} catch (error) {
			console.error(`API Error (${method} ${endpoint}):`, error);
			throw error;
		}
	}
}

export default ApiClient;
