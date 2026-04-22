import { createEffect, createSignal, Show } from "solid-js";

import { type SearchResult, search } from "./api/client";
import ComparisonPanel from "./components/ComparisonPanel";
import ControlPanel from "./components/ControlPanel";
import Results from "./components/Results";
import SearchBar from "./components/SearchBar";
import SegmentedControl from "./components/SegmentedControl";

const SELECTION_STORAGE_KEY = "gh_analyzer_selection";

function loadSelectionFromStorage(): Set<string> {
	if (typeof window === "undefined") {
		return new Set();
	}

	try {
		const raw = window.localStorage.getItem(SELECTION_STORAGE_KEY);
		if (!raw) {
			return new Set();
		}

		const parsed = JSON.parse(raw);
		if (!Array.isArray(parsed)) {
			return new Set();
		}

		const restored = new Set<string>();
		for (const username of parsed) {
			if (typeof username !== "string") {
				return new Set();
			}
			restored.add(username);
		}

		return restored;
	} catch {
		return new Set();
	}
}

export default function App() {
	const [query, setQuery] = createSignal("");
	const [live, setLive] = createSignal(false);
	const [loading, setLoading] = createSignal(false);
	const [error, setError] = createSignal<string | null>(null);
	const [results, setResults] = createSignal<SearchResult[]>([]);
	const [selected, setSelected] = createSignal<Set<string>>(
		loadSelectionFromStorage(),
	);
	const [compareOpen, setCompareOpen] = createSignal(false);

	const visibleSelected = () =>
		results().filter((result) => selected().has(result.username));
	const visibleCount = () => visibleSelected().length;

	const canCompare = () => visibleCount() >= 2 && visibleCount() <= 5;

	createEffect(() => {
		if (typeof window === "undefined") {
			return;
		}

		try {
			const payload = Array.from(selected());
			if (payload.length === 0) {
				window.localStorage.removeItem(SELECTION_STORAGE_KEY);
				return;
			}

			window.localStorage.setItem(
				SELECTION_STORAGE_KEY,
				JSON.stringify(payload),
			);
		} catch {
			// Ignore storage write errors and keep in-memory selection usable.
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

	function removeSelected(username: string) {
		const next = new Set(selected());
		next.delete(username);
		setSelected(next);
	}

	function exportSelectionJSON() {
		const snapshot = visibleSelected();
		const content = JSON.stringify(snapshot, null, 2);
		downloadFile("gh-analyzer-selection.json", content, "application/json");
	}

	function exportSelectionMarkdown() {
		const snapshot = visibleSelected();
		const lines: string[] = [
			"# GH Analyzer Selection",
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

		downloadFile("gh-analyzer-selection.md", lines.join("\n"), "text/markdown");
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
						selected={selected()}
						results={results()}
						onCompare={() => {
							if (!canCompare()) {
								return;
							}
							setCompareOpen(true);
						}}
						onRemove={removeSelected}
						onClear={() => setSelected(new Set())}
						onExportJSON={exportSelectionJSON}
						onExportMD={exportSelectionMarkdown}
					/>

					<section class="flex min-h-0 flex-1 flex-col rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
						<Show when={compareOpen()}>
							<ComparisonPanel
								results={visibleSelected()}
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
								/>
							</Show>
						</div>
					</section>
				</main>
			</div>
		</div>
	);
}
