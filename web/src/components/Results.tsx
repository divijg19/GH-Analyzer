import type { Report } from "../api/client";
import Highlights from "./Highlights";
import Scores from "./Scores";
import TopRepos from "./TopRepos";

type ResultsProps = {
	report: Report;
};

export default function Results(props: ResultsProps) {
	return (
		<article class="space-y-6">
			<header>
				<h2 class="text-xl font-semibold">{props.report.username}</h2>
				<p class="mb-2 mt-6 text-center text-5xl font-bold">
					{props.report.scores.overall}
				</p>
				<p class="text-center text-sm text-gray-500">Overall Score</p>
			</header>

			<section>
				<p class="mt-4 text-gray-600">{props.report.summary}</p>
			</section>

			<Scores scores={props.report.scores} />
			<Highlights highlights={props.report.highlights} />
			<TopRepos repos={props.report.top_repos} />
		</article>
	);
}
