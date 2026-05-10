// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

type Evaluation struct {
	Agent Evaluator `pkl:"agent"`

	Incumbent EvaluationSystemBinding `pkl:"incumbent"`

	Challenger ChallengerEvaluationBinding `pkl:"challenger"`

	Scoring Scoring `pkl:"scoring"`

	Report Report `pkl:"report"`
}
