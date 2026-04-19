import { For } from "solid-js";
import type { SearchResult } from "../api/client";
import CandidateCard from "./CandidateCard";

type ResultsProps = {
	results: SearchResult[];
	selectedSet: Set<string>;
	onToggle: (username: string) => void;
	onAddToShortlist: (result: SearchResult) => void;
	shortlistSet: Map<string, SearchResult>;
};

export default function Results(props: ResultsProps) {
	return (
		<div class="space-y-3">
			<For each={props.results}>
				{(result) => (
					<CandidateCard
						result={result}
						selected={props.selectedSet.has(result.username)}
						onToggle={props.onToggle}
						onAddToShortlist={props.onAddToShortlist}
						shortlisted={props.shortlistSet.has(result.username)}
					/>
				)}
			</For>
		</div>
	);
}
