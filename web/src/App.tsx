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

	const analyze = async () => {
		if (loading()) {
			return;
		}

		const value = username().trim();
		setUsername(value);

		if (!value) {
			setError(formatErrorMessage("missing username"));
			setResult(null);
			return;
		}

		setError(null);
		setResult(null);
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
		<div class="h-screen bg-gray-100 text-gray-900">
			<header class="sticky top-0 border-b bg-white px-6 py-4">
				<div class="flex items-center justify-between">
					<p class="shrink-0 font-semibold">GH Analyzer</p>
					<SearchBar
						username={username()}
						loading={loading()}
						onUsernameChange={setUsername}
						onSubmit={analyze}
					/>
				</div>
			</header>

			<div class="flex h-[calc(100vh-64px)] overflow-hidden">
				<aside class="w-64 border-r bg-white p-4 text-sm text-gray-500">
					<h2 class="mb-2 text-sm font-semibold text-gray-700">Filters</h2>
					<p>Coming soon</p>
				</aside>

				<main class="flex-1 overflow-y-auto p-6">
					<Show when={loading()}>
						<div class="flex h-full items-center justify-center">
							<div class="flex flex-col items-center gap-3 text-gray-600">
								<span class="h-6 w-6 animate-spin rounded-full border-2 border-gray-300 border-t-black" />
								<p>Analyzing "{username()}"...</p>
							</div>
						</div>
					</Show>

					<Show when={!loading() && error()}>
						{(message) => (
							<div class="flex h-full items-center justify-center">
								<p class="text-red-500">❌ {message()}</p>
							</div>
						)}
					</Show>

					<Show when={!loading() && !error() && !result()}>
						<div class="flex h-full items-center justify-center text-gray-500">
							<p>Search for a GitHub user to begin analysis</p>
						</div>
					</Show>

					<Show when={!loading() && result()}>
						{(report) => <Results report={report()} />}
					</Show>
				</main>
			</div>
		</div>
	);
}
