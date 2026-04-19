import { For } from "solid-js";

import type { SearchResult } from "../api/client";
import CandidateCard from "./CandidateCard";

type ResultsProps = {
	results: SearchResult[];
};

export default function Results(props: ResultsProps) {
	return (
		<div class="space-y-3">
			<For each={props.results}>
				{(result) => <CandidateCard result={result} />}
			</For>
		</div>
	);
}
