// Code generated from Pkl module `searchbench.config.Experiment`. DO NOT EDIT.
package generated

type Evaluation struct {
	Agent Evaluator `pkl:"agent"`

	Baseline EvaluationSystemBinding `pkl:"baseline"`

	Candidate CandidateEvaluationBinding `pkl:"candidate"`

	Scoring Scoring `pkl:"scoring"`

	Report Report `pkl:"report"`
}
