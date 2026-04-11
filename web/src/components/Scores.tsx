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
		<section
			style={{
				padding: "16px",
				border: "1px solid #e5e7eb",
				"border-radius": "6px",
			}}
		>
			<h3 style={{ margin: "0 0 12px 0", "font-size": "18px" }}>Scores</h3>
			<div style={{ display: "grid", gap: "10px" }}>
				{rows.map((row) => (
					<div style={{ display: "grid", gap: "6px" }}>
						<div
							style={{
								display: "flex",
								"justify-content": "space-between",
								"font-size": "14px",
							}}
						>
							<span>{row.label}</span>
							<span style={{ "font-weight": "600" }}>{row.value}</span>
						</div>
						<div
							style={{
								height: "8px",
								"background-color": "#e5e7eb",
								"border-radius": "999px",
								overflow: "hidden",
							}}
						>
							<div
								style={{
									height: "100%",
									width: `${Math.max(0, Math.min(100, row.value))}%`,
									"background-color": "#111827",
								}}
							/>
						</div>
					</div>
				))}
			</div>
		</section>
	);
}
