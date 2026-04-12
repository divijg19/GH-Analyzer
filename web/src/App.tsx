import {
	createEffect,
	createMemo,
	createSignal,
	onCleanup,
	Show,
} from "solid-js";
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
	const [loadingMessageIndex, setLoadingMessageIndex] = createSignal(0);
	const [stateVisible, setStateVisible] = createSignal(false);
	let mainRef: HTMLDivElement | undefined;

	const currentState = createMemo(() => {
		if (loading()) return "loading";
		if (error()) return "error";
		if (result()) return "result";
		return "empty";
	});

	const loadingMessage = createMemo(() => {
		const messages = [
			`Analyzing "${username()}"...`,
			"Fetching repositories...",
			"Computing signals...",
		];

		return messages[loadingMessageIndex() % messages.length];
	});

	createEffect(() => {
		if (!loading()) {
			setLoadingMessageIndex(0);
			return;
		}

		const interval = setInterval(() => {
			setLoadingMessageIndex((value) => value + 1);
		}, 700);

		onCleanup(() => clearInterval(interval));
	});

	createEffect(() => {
		currentState();
		setStateVisible(false);

		const frame = requestAnimationFrame(() => {
			setStateVisible(true);
		});

		onCleanup(() => cancelAnimationFrame(frame));
	});

	const analyze = async () => {
		if (loading()) {
			return;
		}

		mainRef?.scrollTo({ top: 0 });

		const value = username().trim();
		setUsername(value);

		if (!value) {
			setError(formatErrorMessage("missing username"));
			setResult(null);
			return;
		}

		setError(null);
		setResult(null);
		setLoadingMessageIndex(0);
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
				<header class="mx-auto mt-2 w-full max-w-6xl shrink-0 rounded-lg bg-white px-6 py-3 shadow-[0_1px_2px_rgba(0,0,0,0.04)]">
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
							class="rounded-md bg-gray-100 px-3 py-1 text-sm text-gray-900"
						>
							Analyze
						</button>
						<button
							type="button"
							class="rounded-md px-3 py-1 text-sm text-gray-500"
						>
							Search
						</button>
					</div>
				</header>

				<div class="mx-auto mt-6 flex min-h-0 w-full max-w-6xl flex-1 gap-6 overflow-hidden">
					<aside class="w-64 self-start rounded-xl border border-gray-200 bg-white p-4 text-sm text-gray-400 shadow-sm transition-colors duration-150 hover:bg-gray-50">
						<p>Filters (coming soon)</p>
					</aside>

					<main
						ref={mainRef}
						class="mx-auto h-full min-h-0 max-w-3xl flex-1 overflow-y-auto rounded-2xl border border-gray-200 bg-white p-8 shadow-sm transition-shadow duration-150 hover:shadow-md"
					>
						<Show when={loading()}>
							<div
								class="flex h-full items-center justify-center py-4 transition-opacity duration-150"
								classList={{
									"opacity-100": stateVisible(),
									"opacity-0": !stateVisible(),
								}}
							>
								<div class="flex flex-col items-center gap-3 text-gray-600">
									<span class="h-6 w-6 animate-spin rounded-full border-2 border-gray-300 border-t-black" />
									<p class="text-sm">{loadingMessage()}</p>
								</div>
							</div>
						</Show>

						<Show when={!loading() && error()}>
							{(message) => (
								<div
									class="flex h-full items-center justify-center py-4 transition-opacity duration-150"
									classList={{
										"opacity-100": stateVisible(),
										"opacity-0": !stateVisible(),
									}}
								>
									<p class="text-sm text-red-500">❌ {message()}</p>
								</div>
							)}
						</Show>

						<Show when={!loading() && !error() && !result()}>
							<div
								class="flex h-full items-center justify-center py-4 text-gray-500 transition-opacity duration-150"
								classList={{
									"opacity-100": stateVisible(),
									"opacity-0": !stateVisible(),
								}}
							>
								<p class="text-sm">
									Search for a GitHub user to begin analysis
								</p>
							</div>
						</Show>

						<Show when={!loading() && result()}>
							{(report) => (
								<div
									class="transition-opacity duration-150"
									classList={{
										"opacity-100": stateVisible(),
										"opacity-0": !stateVisible(),
									}}
								>
									<Results report={report()} />
								</div>
							)}
						</Show>
					</main>
				</div>
			</div>
		</div>
	);
}
