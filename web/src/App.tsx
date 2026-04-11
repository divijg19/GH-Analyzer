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
		<div class="app-shell">
			<aside class="left-panel">
				<div class="left-panel-inner">
					<h1 class="app-title">GH Analyzer</h1>
					<p class="app-subtitle">Analyze GitHub developer signals</p>
					<SearchBar
						username={username()}
						loading={loading()}
						onUsernameChange={setUsername}
						onSubmit={analyze}
					/>
				</div>
			</aside>

			<main class="right-panel">
				<div class="right-panel-inner">
					<Show when={loading()}>
						<div class="state-view">
							<span class="spinner" aria-hidden="true" />
							<p class="state-text">Analyzing "{activeUsername()}"...</p>
						</div>
					</Show>

					<Show when={!loading() && error()}>
						{(message) => (
							<div class="error-card">
								<p class="error-text">❌ {message()}</p>
							</div>
						)}
					</Show>

					<Show when={!loading() && !error() && !result()}>
						<div class="state-view">
							<p class="state-text">Run an analysis to evaluate a developer</p>
						</div>
					</Show>

					<Show when={!loading() && result()}>
						{(report) => <Results report={report()} />}
					</Show>
				</div>
			</main>
		</div>
	);
}
