import type { Report } from "../api/client";
import Highlights from "./Highlights";
import Scores from "./Scores";
import TopRepos from "./TopRepos";

type ResultsProps = {
	report: Report;
};

export default function Results(props: ResultsProps) {
	return (
		<section style={{ margin: "20px 0 0 0", display: "grid", gap: "12px" }}>
			<div
				style={{
					padding: "16px",
					border: "1px solid #e5e7eb",
					"border-radius": "6px",
				}}
			>
				<h2 style={{ margin: "0 0 8px 0", "font-size": "22px" }}>
					{props.report.username}
				</h2>
				<p style={{ margin: "0", color: "#4b5563" }}>{props.report.summary}</p>
			</div>

			<Scores scores={props.report.scores} />
			<Highlights highlights={props.report.highlights} />
			<TopRepos repos={props.report.top_repos} />
		</section>
	);
}
