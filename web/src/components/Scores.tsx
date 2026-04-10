import type { Report } from "../api/client";

type ScoresProps = {
	scores: Report["scores"];
};

export default function Scores(props: ScoresProps) {
	return (
		<section
			style={{
				padding: "16px 0",
				border: "1px solid #e5e7eb",
				"border-radius": "6px",
			}}
		>
			<h3 style={{ margin: "0 16px 12px 16px", "font-size": "18px" }}>
				Scores
			</h3>
			<dl
				style={{
					margin: "0",
					padding: "0 16px 8px 16px",
					display: "grid",
					"grid-template-columns": "1fr auto",
					gap: "10px 12px",
				}}
			>
				<dt>Ownership</dt>
				<dd style={{ margin: "0" }}>{props.scores.ownership}</dd>
				<dt>Consistency</dt>
				<dd style={{ margin: "0" }}>{props.scores.consistency}</dd>
				<dt>Depth</dt>
				<dd style={{ margin: "0" }}>{props.scores.depth}</dd>
				<dt>Overall</dt>
				<dd style={{ margin: "0", "font-weight": "600" }}>
					{props.scores.overall}
				</dd>
			</dl>
		</section>
	);
}
