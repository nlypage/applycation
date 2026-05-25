import { useEffect, useState } from "react";

type HealthState =
	| { kind: "loading" }
	| { kind: "healthy"; status: string }
	| { kind: "error"; message: string };

export function App() {
	const [healthState, setHealthState] = useState<HealthState>({
		kind: "loading",
	});

	useEffect(() => {
		const controller = new AbortController();

		const loadHealth = async () => {
			try {
				const response = await fetch("/health", {
					headers: { Accept: "application/json" },
					signal: controller.signal,
				});

				if (!response.ok) {
					setHealthState({
						kind: "error",
						message: `Backend returned ${response.status}`,
					});
					return;
				}

				const payload: unknown = await response.json();
				const status =
					typeof payload === "object" &&
					payload !== null &&
					"status" in payload &&
					typeof payload.status === "string"
						? payload.status
						: "unknown";

				setHealthState({ kind: "healthy", status });
			} catch (error) {
				if (error instanceof DOMException && error.name === "AbortError") {
					return;
				}

				setHealthState({
					kind: "error",
					message: "Failed to reach backend health endpoint",
				});
			}
		};

		void loadHealth();

		return () => controller.abort();
	}, []);

	return (
		<main>
			<h1>applycation</h1>
			<p>Phase 1 backend integration status:</p>
			{healthState.kind === "loading" && <p>Checking backend health...</p>}
			{healthState.kind === "healthy" && (
				<p>Backend health: {healthState.status}</p>
			)}
			{healthState.kind === "error" && (
				<p>Backend health error: {healthState.message}</p>
			)}
		</main>
	);
}
