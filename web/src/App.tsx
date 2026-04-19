import { createSignal, Show } from "solid-js";

import { type SearchResult, search } from "./api/client";
import Results from "./components/Results";
import SearchBar from "./components/SearchBar";

export default function App() {
	const [query, setQuery] = createSignal("");
	const [live, setLive] = createSignal(false);
	const [loading, setLoading] = createSignal(false);
	const [error, setError] = createSignal<string | null>(null);
	const [results, setResults] = createSignal<SearchResult[]>([]);

	const handleSearch = async (value: string) => {
		const trimmed = value.trim();
		setQuery(trimmed);

		if (!trimmed) {
			setResults([]);
			setError(null);
			return;
		}

		setLoading(true);
		setError(null);

		try {
			const payload = await search(trimmed, live());
			setResults(payload.results);
		} catch {
			setError("Failed to fetch data");
			setResults([]);
		} finally {
			setLoading(false);
		}
	};

	return (
		<div class="h-screen overflow-hidden bg-slate-50 text-slate-900">
			<div class="mx-auto flex h-full max-w-5xl flex-col px-4 py-6">
				<header class="mx-auto w-full max-w-4xl rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
					<div class="mb-3 flex items-center justify-between">
						<div>
							<p class="text-sm font-semibold uppercase tracking-wide text-slate-500">
								GH Analyzer
							</p>
							<h1 class="text-xl font-semibold text-slate-900">
								Developer Search
							</h1>
						</div>
						<p class="text-sm text-slate-500">
							Mode: {live() ? "Live" : "Dataset"}
						</p>
					</div>

					<SearchBar
						onSearch={handleSearch}
						live={live()}
						onToggleLive={() => setLive((value) => !value)}
					/>
				</header>

				<main class="mx-auto mt-5 flex min-h-0 w-full max-w-4xl flex-1">
					<section class="min-h-0 w-full overflow-y-auto rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
						<Show when={query().length > 0}>
							<p class="mb-4 text-sm text-slate-500">Query: {query()}</p>
						</Show>

						<Show when={loading()}>
							<div class="flex h-full items-center justify-center text-sm text-slate-600">
								Searching...
							</div>
						</Show>

						<Show when={!loading() && error()}>
							<div class="flex h-full items-center justify-center text-sm text-red-600">
								{error()}
							</div>
						</Show>

						<Show when={!loading() && !error() && results().length === 0}>
							<div class="flex h-full items-center justify-center text-sm text-slate-600">
								No candidates found
							</div>
						</Show>

						<Show when={!loading() && !error() && results().length > 0}>
							<Results results={results()} />
						</Show>
					</section>
				</main>
			</div>
		</div>
	);
}
