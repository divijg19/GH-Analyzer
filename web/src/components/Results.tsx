import type { Report } from "../api/client";
import Highlights from "./Highlights";
import Scores from "./Scores";
import TopRepos from "./TopRepos";

type ResultsProps = {
	report: Report;
};

export default function Results(props: ResultsProps) {
	return (
		<article class="mx-auto max-w-2xl space-y-6 rounded-lg border bg-white p-6">
			<header>
				<h2 class="text-xl font-semibold">{props.report.username}</h2>
				<p class="mt-4 text-center text-sm text-gray-500">Overall Score</p>
				<p class="text-center text-5xl font-bold">
					{props.report.scores.overall}
				</p>
			</header>

			<section>
				<h3 class="text-sm font-semibold">Summary</h3>
				<p class="mt-2 text-gray-600">{props.report.summary}</p>
			</section>

			<Scores scores={props.report.scores} />
			<Highlights highlights={props.report.highlights} />
			<TopRepos repos={props.report.top_repos} />
		</article>
	);
}
