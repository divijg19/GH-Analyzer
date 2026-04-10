import { createSignal, Show } from "solid-js";
import { fetchReport, type Report } from "./api/client";
import Results from "./components/Results";
import SearchBar from "./components/SearchBar";

export default function App() {
	const [username, setUsername] = createSignal("");
	const [loading, setLoading] = createSignal(false);
	const [error, setError] = createSignal<string | null>(null);
	const [result, setResult] = createSignal<Report | null>(null);

	const analyze = async () => {
		const value = username().trim();
		setUsername(value);

		if (!value) {
			setError("missing username");
			setResult(null);
			return;
		}

		setLoading(true);
		setError(null);

		try {
			const report = await fetchReport(value);
			setResult(report);
		} catch (err) {
			const message = err instanceof Error ? err.message : "request failed";
			setError(message);
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
					"max-width": "760px",
					padding: "24px",
					"background-color": "#ffffff",
					border: "1px solid #e5e7eb",
					"border-radius": "8px",
				}}
			>
				<h1 style={{ margin: "0 0 8px 0", "font-size": "28px" }}>
					GH-Analyzer
				</h1>
				<p style={{ margin: "0 0 20px 0", color: "#4b5563" }}>
					Analyze a GitHub username using the local gh-analyzer API.
				</p>

				<SearchBar
					username={username()}
					loading={loading()}
					onUsernameChange={setUsername}
					onSubmit={analyze}
				/>

				<Show when={loading()}>
					<p style={{ margin: "16px 0 0 0", color: "#374151" }}>Analyzing...</p>
				</Show>

				<Show when={error()}>
					{(message) => (
						<p style={{ margin: "16px 0 0 0", color: "#b91c1c" }}>
							{message()}
						</p>
					)}
				</Show>

				<Show when={result()}>{(report) => <Results report={report()} />}</Show>
			</main>
		</div>
	);
}
