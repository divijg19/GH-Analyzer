import type { Report } from "../api/client";
import Highlights from "./Highlights";
import Scores from "./Scores";
import TopRepos from "./TopRepos";

type ResultsProps = {
	report: Report;
};

export default function Results(props: ResultsProps) {
	return (
		<section style={{ margin: "8px 0 0 0", display: "grid", gap: "12px" }}>
			<div
				style={{
					padding: "16px",
					border: "1px solid #e5e7eb",
					"border-radius": "6px",
					display: "grid",
					gap: "8px",
				}}
			>
				<h2 style={{ margin: "0", "font-size": "24px" }}>
					{props.report.username}
				</h2>
				<div>
					<p
						style={{
							margin: "0",
							"font-size": "12px",
							"font-weight": "600",
							color: "#6b7280",
							"text-transform": "uppercase",
							"letter-spacing": "0.04em",
						}}
					>
						Overall Score
					</p>
					<p
						style={{
							margin: "2px 0 0 0",
							"font-size": "34px",
							"font-weight": "700",
						}}
					>
						{props.report.scores.overall}
					</p>
				</div>
			</div>

			<Scores scores={props.report.scores} />

			<section
				style={{
					padding: "16px",
					border: "1px solid #e5e7eb",
					"border-radius": "6px",
				}}
			>
				<h3 style={{ margin: "0 0 8px 0", "font-size": "18px" }}>Summary</h3>
				<p style={{ margin: "0", color: "#4b5563" }}>{props.report.summary}</p>
			</section>

			<Highlights highlights={props.report.highlights} />
			<TopRepos repos={props.report.top_repos} />
		</section>
	);
}
