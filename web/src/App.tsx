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
		<div class="h-screen overflow-hidden bg-gray-50 text-gray-900">
			<div class="mx-auto flex h-full box-border flex-col px-6 py-6">
				<header class="mx-auto mt-2 w-full max-w-4xl shrink-0 rounded-xl border bg-white px-6 py-4 shadow-sm">
					<div class="flex items-center">
						<p class="shrink-0 text-sm font-semibold">GH Analyzer</p>
						<SearchBar
							username={username()}
							loading={loading()}
							onUsernameChange={setUsername}
							onSubmit={analyze}
						/>
					</div>
					<div class="mt-3 flex items-center gap-2">
						<button
							type="button"
							class="rounded-md border bg-gray-100 px-3 py-1 text-sm text-gray-900"
						>
							Analyze
						</button>
						<button
							type="button"
							class="rounded-md border border-transparent px-3 py-1 text-sm text-gray-500"
						>
							Search
						</button>
					</div>
				</header>

				<div class="mx-auto mt-6 flex min-h-0 w-full max-w-6xl flex-1 gap-6 overflow-hidden">
					<aside class="w-64 self-start rounded-xl border bg-white p-4 text-sm text-gray-500 shadow-sm">
						<p>Filters (coming soon)</p>
					</aside>

					<main class="h-full min-h-0 max-w-3xl flex-1 overflow-y-auto rounded-2xl border bg-white p-6 shadow-sm">
						<Show when={loading()}>
							<div class="flex min-h-96 items-center justify-center">
								<div class="flex flex-col items-center gap-3 text-gray-600">
									<span class="h-6 w-6 animate-spin rounded-full border-2 border-gray-300 border-t-black" />
									<p>Analyzing "{username()}"...</p>
								</div>
							</div>
						</Show>

						<Show when={!loading() && error()}>
							{(message) => (
								<div class="flex min-h-96 items-center justify-center">
									<p class="text-red-500">❌ {message()}</p>
								</div>
							)}
						</Show>

						<Show when={!loading() && !error() && !result()}>
							<div class="flex min-h-96 items-center justify-center text-gray-500">
								<p>Search for a GitHub user to begin analysis</p>
							</div>
						</Show>

						<Show when={!loading() && result()}>
							{(report) => <Results report={report()} />}
						</Show>
					</main>
				</div>
			</div>
		</div>
	);
}
