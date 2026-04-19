import { createSignal, Show } from "solid-js";

import type { SearchResult } from "../api/client";

type CandidateCardProps = {
	result: SearchResult;
};

export default function CandidateCard(props: CandidateCardProps) {
	const [expanded, setExpanded] = createSignal(false);

	return (
		<article class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm transition hover:shadow-md">
			<header class="flex flex-wrap items-center justify-between gap-3">
				<div>
					<h3 class="text-lg font-semibold text-slate-900">
						{props.result.username}
					</h3>
					<p class="text-sm text-slate-600">
						Better than {Math.round(props.result.score * 100)}% of candidates
					</p>
				</div>

				<div class="flex items-center gap-2 text-sm">
					<span class="rounded-md border border-slate-200 bg-slate-50 px-2 py-1 font-medium text-slate-700">
						{props.result.score.toFixed(2)}
					</span>
					<span class="rounded-md border border-slate-200 bg-slate-50 px-2 py-1 text-slate-700">
						{props.result.confidence}
					</span>
				</div>
			</header>

			<ul class="mt-3 list-disc space-y-1 pl-5 text-sm text-slate-700">
				{props.result.reasons.slice(0, 3).map((reason) => (
					<li>{reason}</li>
				))}
			</ul>

			<button
				type="button"
				onClick={() => setExpanded((value) => !value)}
				class="mt-4 rounded-md border border-slate-200 bg-slate-50 px-3 py-1.5 text-sm text-slate-700 transition hover:bg-slate-100"
			>
				{expanded() ? "Hide signals" : "Show signals"}
			</button>

			<Show when={expanded()}>
				<div class="mt-3 grid gap-1 text-sm text-slate-700">
					<p>Consistency: {props.result.signals.consistency.toFixed(2)}</p>
					<p>Ownership: {props.result.signals.ownership.toFixed(2)}</p>
					<p>Depth: {props.result.signals.depth.toFixed(2)}</p>
					<p>Activity: {props.result.signals.activity.toFixed(2)}</p>
				</div>
			</Show>
		</article>
	);
}
