package sources

type DummySource struct{}

func (d *DummySource) GetText() (string, error) {
	var dummyText string = "Thise is a Dummy Text! To set up LLM for GoTouch, first sign up at https://console.anthropic.com and generate your API key. Next, save the key to a config file by using the command echo "your-api-key-here" > ~/.config/gotouch/api-key, or set it as an environment variable with export ANTHROPIC_API_KEY="your-api-key-here". Then, edit the config.yaml file to set the text source to LLM by changing it to text: source: llm. Finally, start GoTouch by running gotouch in your terminal."
	return dummyText, nil
}
