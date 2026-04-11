import type { Report } from "../api/client";
import Highlights from "./Highlights";
import Scores from "./Scores";
import TopRepos from "./TopRepos";

type ResultsProps = {
	report: Report;
};

export default function Results(props: ResultsProps) {
	return (
		<article class="candidate-card">
			<header class="candidate-header">
				<h2 class="candidate-name">{props.report.username}</h2>
				<p class="overall-label">Overall Score</p>
				<p class="overall-score">{props.report.scores.overall}</p>
			</header>

			<section class="card-section">
				<h3 class="section-title">Summary</h3>
				<p class="section-text">{props.report.summary}</p>
			</section>

			<Scores scores={props.report.scores} />
			<Highlights highlights={props.report.highlights} />
			<TopRepos repos={props.report.top_repos} />
		</article>
	);
}
