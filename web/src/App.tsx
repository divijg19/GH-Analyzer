import { createEffect, createSignal, Show } from "solid-js";

import { type SearchResult, search } from "./api/client";
import ComparisonPanel from "./components/ComparisonPanel";
import ControlPanel from "./components/ControlPanel";
import Results from "./components/Results";
import SearchBar from "./components/SearchBar";
import SegmentedControl from "./components/SegmentedControl";

const SHORTLIST_STORAGE_KEY = "gh_analyzer_shortlist";

function isStoredSearchResult(value: unknown): value is SearchResult {
	if (typeof value !== "object" || value === null) {
		return false;
	}

	const candidate = value as { username?: unknown };
	return typeof candidate.username === "string";
}

function loadShortlistFromStorage(): Map<string, SearchResult> {
	if (typeof window === "undefined") {
		return new Map();
	}

	try {
		const raw = window.localStorage.getItem(SHORTLIST_STORAGE_KEY);
		if (!raw) {
			return new Map();
		}

		const parsed = JSON.parse(raw);
		if (!Array.isArray(parsed)) {
			return new Map();
		}

		const restored = new Map<string, SearchResult>();
		for (const entry of parsed) {
			if (!isStoredSearchResult(entry)) {
				return new Map();
			}
			restored.set(entry.username, entry);
		}

		return restored;
	} catch {
		return new Map();
	}
}

export default function App() {
	const [query, setQuery] = createSignal("");
	const [live, setLive] = createSignal(false);
	const [loading, setLoading] = createSignal(false);
	const [error, setError] = createSignal<string | null>(null);
	const [results, setResults] = createSignal<SearchResult[]>([]);
	const [selected, setSelected] = createSignal<Set<string>>(new Set());
	const [shortlist, setShortlist] = createSignal<Map<string, SearchResult>>(
		loadShortlistFromStorage(),
	);
	const [compareOpen, setCompareOpen] = createSignal(false);

	const selectedResults = () =>
		results().filter((result) => selected().has(result.username));
	const shortlistResults = () => Array.from(shortlist().values());

	const canCompare = () => selected().size >= 2 && selected().size <= 5;

	createEffect(() => {
		if (typeof window === "undefined") {
			return;
		}

		try {
			const payload = shortlistResults();
			if (payload.length === 0) {
				window.localStorage.removeItem(SHORTLIST_STORAGE_KEY);
				return;
			}

			window.localStorage.setItem(
				SHORTLIST_STORAGE_KEY,
				JSON.stringify(payload),
			);
		} catch {
			// Ignore storage write errors and keep in-memory shortlist usable.
		}
	});

	function toggleSelect(username: string) {
		setSelected((current) => {
			const next = new Set(current);
			if (next.has(username)) {
				next.delete(username);
			} else {
				next.add(username);
			}

			return next;
		});
	}

	function addToShortlist(result: SearchResult) {
		setShortlist((current) => {
			const next = new Map(current);
			next.set(result.username, result);
			return next;
		});
	}

	function removeFromShortlist(username: string) {
		setShortlist((current) => {
			const next = new Map(current);
			next.delete(username);
			return next;
		});
	}

	function addSelectedToShortlist() {
		setShortlist((current) => {
			const next = new Map(current);
			for (const result of selectedResults()) {
				next.set(result.username, result);
			}
			return next;
		});
	}

	function clearShortlist() {
		setShortlist(new Map());
		if (typeof window !== "undefined") {
			try {
				window.localStorage.removeItem(SHORTLIST_STORAGE_KEY);
			} catch {
				// Ignore storage clear errors.
			}
		}
	}

	function exportShortlistJSON() {
		const snapshot = shortlistResults();
		const content = JSON.stringify(snapshot, null, 2);
		downloadFile("gh-analyzer-shortlist.json", content, "application/json");
	}

	function exportShortlistMarkdown() {
		const snapshot = shortlistResults();
		const lines: string[] = [
			"# GH Analyzer Shortlist",
			"",
			"## Candidates",
			"",
		];

		snapshot.forEach((result, index) => {
			lines.push(
				`${index + 1}. ${result.username} — ${result.score.toFixed(2)}`,
			);
			for (const reason of result.reasons) {
				lines.push(`   - ${reason}`);
			}
			if (index < snapshot.length - 1) {
				lines.push("");
			}
		});

		downloadFile("gh-analyzer-shortlist.md", lines.join("\n"), "text/markdown");
	}

	function downloadFile(filename: string, content: string, mimeType: string) {
		if (typeof window === "undefined") {
			return;
		}

		const blob = new Blob([content], { type: `${mimeType};charset=utf-8` });
		const url = window.URL.createObjectURL(blob);
		const anchor = document.createElement("a");
		anchor.href = url;
		anchor.download = filename;
		document.body.appendChild(anchor);
		anchor.click();
		anchor.remove();
		window.URL.revokeObjectURL(url);
	}

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
			<div class="mx-auto flex h-full max-w-7xl flex-col px-4 py-6">
				<header class="w-full rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
					<div class="flex items-center gap-4">
						<div>
							<p class="text-sm font-semibold uppercase tracking-wide text-slate-500">
								GH Analyzer
							</p>
							<h1 class="text-xl font-semibold text-slate-900">
								Developer Search
							</h1>
						</div>
						<div class="min-w-0 flex-1">
							<SearchBar onSearch={handleSearch} loading={loading()} />
						</div>

						<SegmentedControl
							mode={live() ? "live" : "dataset"}
							onChange={(mode) => setLive(mode === "live")}
						/>
					</div>
				</header>

				<main class="mt-5 flex min-h-0 flex-1 gap-4">
					<ControlPanel
						selectedCount={selected().size}
						canCompare={canCompare()}
						onCompare={() => {
							if (!canCompare()) {
								return;
							}
							setCompareOpen(true);
						}}
						onClearSelection={() => setSelected(new Set())}
						canAddSelected={selectedResults().length > 0}
						onAddSelected={addSelectedToShortlist}
						shortlist={shortlistResults()}
						onRemoveShortlist={removeFromShortlist}
						onClearShortlist={clearShortlist}
						onExportJSON={exportShortlistJSON}
						onExportMarkdown={exportShortlistMarkdown}
					/>

					<section class="flex min-h-0 flex-1 flex-col rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
						<Show when={compareOpen()}>
							<ComparisonPanel
								results={selectedResults()}
								onClose={() => setCompareOpen(false)}
							/>
						</Show>

						<Show when={query().length > 0}>
							<p class="mb-4 text-sm text-slate-500">Query: {query()}</p>
						</Show>

						<div class="min-h-0 flex-1 overflow-y-auto pr-1">
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
								<Results
									results={results()}
									selectedSet={selected()}
									onToggle={toggleSelect}
									onAddToShortlist={addToShortlist}
									shortlistSet={shortlist()}
								/>
							</Show>
						</div>
					</section>
				</main>
			</div>
		</div>
	);
}
