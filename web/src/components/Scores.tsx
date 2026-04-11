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
		<section class="card-section">
			<h3 class="section-title">Scores</h3>
			<div class="score-list">
				{rows.map((row) => (
					<div class="score-row">
						<div class="score-row-head">
							<span>{row.label}</span>
							<span class="score-value">{row.value}</span>
						</div>
						<div class="score-track">
							<div
								class="score-fill"
								style={{ width: `${Math.max(0, Math.min(100, row.value))}%` }}
							/>
						</div>
					</div>
				))}
			</div>
		</section>
	);
}
