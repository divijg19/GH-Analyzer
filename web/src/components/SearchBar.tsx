import type { JSX } from "solid-js";

type SearchBarProps = {
	username: string;
	loading: boolean;
	onUsernameChange: (value: string) => void;
	onSubmit: () => void;
};

export default function SearchBar(props: SearchBarProps) {
	const handleSubmit: JSX.EventHandler<HTMLFormElement, SubmitEvent> = (
		event,
	) => {
		event.preventDefault();
		props.onSubmit();
	};

	return (
		<form
			onSubmit={handleSubmit}
			style={{ display: "flex", gap: "8px", "align-items": "center" }}
		>
			<input
				type="text"
				placeholder="GitHub username"
				value={props.username}
				onInput={(event) => props.onUsernameChange(event.currentTarget.value)}
				disabled={props.loading}
				style={{
					flex: "1",
					padding: "10px 12px",
					border: "1px solid #d1d5db",
					"border-radius": "6px",
					"font-size": "14px",
				}}
			/>
			<button
				type="submit"
				disabled={props.loading}
				style={{
					padding: "10px 14px",
					border: "1px solid #111827",
					"border-radius": "6px",
					"background-color": "#111827",
					color: "#ffffff",
					"font-size": "14px",
					cursor: props.loading ? "not-allowed" : "pointer",
					opacity: props.loading ? "0.7" : "1",
				}}
			>
				Analyze
			</button>
		</form>
	);
}
