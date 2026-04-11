import type { Report } from "../api/client";

type ScoresProps = {
	scores: Report["scores"];
};

export default function Scores(props: ScoresProps) {
	const rows = [
		{ label: "Ownership", value: props.scores.ownership },
		{ label: "Consistency", value: props.scores.consistency },
		{ label: "Depth", value: props.scores.depth },
		{ label: "Overall", value: props.scores.overall },
	];

	return (
		<section>
			<h3 class="text-sm font-semibold">Scores</h3>
			<div class="mt-2 space-y-3">
				{rows.map((row) => (
					<div>
						<div class="flex items-center justify-between text-sm">
							<span>{row.label}</span>
							<span>{row.value}</span>
						</div>
						<div class="mx-2 mt-1 h-2 w-full rounded bg-gray-200">
							<div
								class="h-2 rounded bg-black"
								style={{ width: `${Math.max(0, Math.min(100, row.value))}%` }}
							/>
						</div>
					</div>
				))}
			</div>
		</section>
	);
}
