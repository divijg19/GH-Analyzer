import { For } from "solid-js";

import type { SearchResult } from "../api/client";

type ComparisonPanelProps = {
	results: SearchResult[];
	onClose: () => void;
};

export default function ComparisonPanel(props: ComparisonPanelProps) {
	const gridTemplateColumns = () =>
		`minmax(120px, 140px) repeat(${props.results.length}, minmax(0, 1fr))`;

	const maxScore = () =>
		props.results.length > 0
			? Math.max(...props.results.map((result) => result.score))
			: 0;

	const maxSignal = (signal: keyof SearchResult["signals"]) =>
		props.results.length > 0
			? Math.max(...props.results.map((result) => result.signals[signal]))
			: 0;

	const highlightClass =
		"rounded-md border border-sky-200 bg-sky-50 font-semibold text-sky-900";
	const baseClass = "rounded-md border border-transparent";

	return (
		<div class="mb-4 rounded-xl border border-slate-200 bg-slate-50 p-4">
			<div class="mb-3 flex items-center justify-between">
				<h2 class="text-sm font-semibold text-slate-700">Comparison</h2>
				<button
					type="button"
					onClick={props.onClose}
					class="rounded-md border border-slate-200 bg-white px-2 py-1 text-xs text-slate-600 hover:bg-slate-50"
				>
					Close
				</button>
			</div>

			<div class="overflow-x-auto">
				<div
					class="grid min-w-160 gap-2 text-center text-sm"
					style={{ "grid-template-columns": gridTemplateColumns() }}
				>
					<div class="text-left font-medium text-slate-600">Username</div>
					<For each={props.results}>
						{(result) => (
							<div class="font-semibold text-slate-800">{result.username}</div>
						)}
					</For>

					<div class="text-left font-medium text-slate-600">Score</div>
					<For each={props.results}>
						{(result) => (
							<div
								class={result.score === maxScore() ? highlightClass : baseClass}
							>
								<div>{result.score.toFixed(2)}</div>
								<div class="text-xs text-slate-500">
									{Math.round(result.score * 100)}%
								</div>
							</div>
						)}
					</For>

					<div class="text-left font-medium text-slate-600">Consistency</div>
					<For each={props.results}>
						{(result) => (
							<div
								class={
									result.signals.consistency === maxSignal("consistency")
										? highlightClass
										: baseClass
								}
							>
								{result.signals.consistency.toFixed(2)}
							</div>
						)}
					</For>

					<div class="text-left font-medium text-slate-600">Ownership</div>
					<For each={props.results}>
						{(result) => (
							<div
								class={
									result.signals.ownership === maxSignal("ownership")
										? highlightClass
										: baseClass
								}
							>
								{result.signals.ownership.toFixed(2)}
							</div>
						)}
					</For>

					<div class="text-left font-medium text-slate-600">Depth</div>
					<For each={props.results}>
						{(result) => (
							<div
								class={
									result.signals.depth === maxSignal("depth")
										? highlightClass
										: baseClass
								}
							>
								{result.signals.depth.toFixed(2)}
							</div>
						)}
					</For>

					<div class="text-left font-medium text-slate-600">Activity</div>
					<For each={props.results}>
						{(result) => (
							<div
								class={
									result.signals.activity === maxSignal("activity")
										? highlightClass
										: baseClass
								}
							>
								{result.signals.activity.toFixed(2)}
							</div>
						)}
					</For>
				</div>
			</div>
		</div>
	);
}
