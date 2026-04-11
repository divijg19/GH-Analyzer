import { createSignal, Show } from "solid-js";
import { fetchReport, type Report } from "./api/client";
import Results from "./components/Results";
import SearchBar from "./components/SearchBar";

function formatErrorMessage(message: string): string {
	if (message === "missing username") {
		return "Please enter a GitHub username.";
	}
	if (message === "invalid API response") {
		return "Received an invalid response from the analyzer API.";
	}

	return message;
}

export default function App() {
	const [username, setUsername] = createSignal("");
	const [loading, setLoading] = createSignal(false);
	const [error, setError] = createSignal<string | null>(null);
	const [result, setResult] = createSignal<Report | null>(null);
	const [activeUsername, setActiveUsername] = createSignal("");

	const analyze = async () => {
		const value = username().trim();
		setUsername(value);

		if (!value) {
			setError(formatErrorMessage("missing username"));
			setResult(null);
			return;
		}

		setError(null);
		setResult(null);
		setActiveUsername(value);
		setLoading(true);

		try {
			const report = await fetchReport(value);
			setResult(report);
		} catch (err) {
			if (err instanceof TypeError) {
				setError(
					"Cannot reach analyzer API. Ensure backend server is running on :8080.",
				);
			} else {
				const message =
					err instanceof Error ? err.message : "analysis request failed";
				setError(formatErrorMessage(message));
			}
			setResult(null);
		} finally {
			setLoading(false);
		}
	};

	return (
		<div
			style={{
				"min-height": "100vh",
				padding: "32px 16px",
				"background-color": "#f5f6f8",
				"font-family":
					"system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial, sans-serif",
				color: "#1f2937",
			}}
		>
			<main
				style={{
					margin: "0 auto",
					"max-width": "700px",
					padding: "24px",
					"background-color": "#ffffff",
					border: "1px solid #e5e7eb",
					"border-radius": "8px",
					display: "grid",
					gap: "16px",
				}}
			>
				<style>
					{`@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }`}
				</style>
				<h1 style={{ margin: "0 0 8px 0", "font-size": "28px" }}>
					GH-Analyzer
				</h1>
				<p style={{ margin: "0", color: "#4b5563" }}>
					Analyze a GitHub username using the local gh-analyzer API.
				</p>

				<SearchBar
					username={username()}
					loading={loading()}
					onUsernameChange={setUsername}
					onSubmit={analyze}
				/>

				<Show when={loading()}>
					<div style={{ display: "flex", "align-items": "center", gap: "8px" }}>
						<span
							aria-hidden="true"
							style={{
								width: "14px",
								height: "14px",
								border: "2px solid #d1d5db",
								"border-top-color": "#111827",
								"border-radius": "999px",
								animation: "spin 0.8s linear infinite",
							}}
						/>
						<p style={{ margin: "0", color: "#374151" }}>
							Analyzing {activeUsername()}...
						</p>
					</div>
				</Show>

				<Show when={error()}>
					{(message) => (
						<p style={{ margin: "0", color: "#b91c1c", "font-weight": "500" }}>
							❌ {message()}
						</p>
					)}
				</Show>

				<Show when={!loading() && !error() && !result()}>
					<p style={{ margin: "0", color: "#4b5563", "text-align": "center" }}>
						Enter a GitHub username to analyze developer signals
					</p>
				</Show>

				<Show when={result()}>{(report) => <Results report={report()} />}</Show>
			</main>
		</div>
	);
}
