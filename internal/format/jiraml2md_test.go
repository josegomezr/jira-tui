package format

import (
	"testing"
	"strings"
)

func TestBold(t *testing.T) {
	expectations := []struct {
		input string
		expectedOutput string
	}{
		{
			"a *bold* text\n",
			"a **bold** text\n",
		},
		{
			"a {*}bold{*} text\n",
			"a **bold** text\n",
		},
	}
	for _, expectation := range expectations {
		output, err := NewConverter(strings.NewReader(expectation.input)).Convert()
		if err != nil {
			t.Errorf("error: %s", err)
		}
		if output != expectation.expectedOutput {
			t.Errorf("Failed conversion.\nExpected: %q\nGot: %q", expectation.expectedOutput, output)
		}
	}
}

func TestItalics(t *testing.T) {
	expectations := []struct {
		input string
		expectedOutput string
	}{
		{
			"one _italic_ s\n",
			"one __italic__ s\n",
		},
	}

	for _, expectation := range expectations {
		output, err := NewConverter(strings.NewReader(expectation.input)).Convert()
		if err != nil {
			t.Errorf("error: %s", err)
		}
		if output != expectation.expectedOutput {
			t.Errorf("Failed conversion.\nExpected: %q\nGot: %q", expectation.expectedOutput, output)
		}
	}
}

func TestBolditalics(t *testing.T) {
	expectations := []struct {
		input string
		expectedOutput string
	}{
		{
			"*_bold-italics_*\n",
			"__**bold-italics**__\n",
		},
	}

	for _, expectation := range expectations {
		output, err := NewConverter(strings.NewReader(expectation.input)).Convert()
		if err != nil {
			t.Errorf("error: %s", err)
		}
		if output != expectation.expectedOutput {
			t.Errorf("Failed conversion.\nExpected: %q\nGot: %q", expectation.expectedOutput, output)
		}
	}
}

func TestStripColors(t *testing.T) {
	expectations := []struct {
		input string
		expectedOutput string
	}{
		{
			"{color:#cafeca}da text{color}\n",
			"da text\n",
		},
		{
			"{color}da text{color}\n",
			"da text\n",
		},
		{
			"da text{color}\n",
			"da text\n",
		},
	}

	for _, expectation := range expectations {
		output, err := NewConverter(strings.NewReader(expectation.input)).Convert()
		if err != nil {
			t.Errorf("error: %s", err)
		}
		if output != expectation.expectedOutput {
			t.Errorf("Failed conversion.\nExpected: %q\nGot: %q", expectation.expectedOutput, output)
		}
	}
}

func TestExtractLink(t *testing.T) {
	expectations := []struct {
		input string
		expectedOutput string
	}{
		{
			"[https://foo.bar/]\n",
			"[link-0][0]\n\n--\n[0]: https://foo.bar/\n",
		},
		{
			"[baz|https://foo.bar/]\n",
			"[baz][0]\n\n--\n[0]: https://foo.bar/ <baz>\n",
		},
		{
			"Something [here|https://foo.bar/] and here.\n",
			"Something [here][0] and here.\n\n--\n[0]: https://foo.bar/ <here>\n",
		},
		{
			"A bare link https://da.link/ and then a labeled link [here|https://foo.bar/] and here.\n",
			"A bare link [link-0][0] and then a labeled link [here][1] and here.\n\n--\n[0]: https://da.link/\n[1]: https://foo.bar/ <here>\n",
		},
		{
			"A bare link https://da.link/ and then a labeled link with highlights [https://link.to/place/with/highligth#:~:text=foo bar baz|https://link.to/place/with/highligth#:~:text=foo%20bar%20baz]\n",
			"A bare link [link-0][0] and then a labeled link with highlights [link-1][1]\n\n--\n[0]: https://da.link/\n[1]: https://link.to/place/with/highligth#:~:text=foo%20bar%20baz <link-1>\n",
		},		{
			"A bare link https://da.link/ and then a mention [~username]\n",
			"A bare link [link-0][0] and then a mention @[~username]\n\n--\n[0]: https://da.link/\n",
		},
	}

	for idx, expectation := range expectations {
		output, err := NewConverter(strings.NewReader(expectation.input)).Convert()
		if err != nil {
			t.Errorf("error: %s", err)
		}
		if output != expectation.expectedOutput {
			t.Errorf("[%d] Failed conversion.\nInput:\n%s\nExpected:\n%s\nGot:\n%s", idx, expectation.input, expectation.expectedOutput, output)
			return
		}
	}
}

func TestExtractImage(t *testing.T) {
	expectations := []struct {
		input string
		expectedOutput string
	}{
		{
			"!https://foo.bar/!\n",
			"![][image-0]\n\n--\n[image-0]: https://foo.bar/\n",
		},
		{
			"!https://foo.bar/|alt=baz!\n",
			"![baz][image-0]\n\n--\n[image-0]: https://foo.bar/\n",
		},
		{
			"Something !https://foo.bar/|alt=here! and here.\n",
			"Something ![here][image-0] and here.\n\n--\n[image-0]: https://foo.bar/\n",
		},
		{
			"Something !https://foo.bar/|alt=here! and here with https://a.link/.\n",
			"Something ![here][image-0] and here with [link-1][1]\n\n--\n[image-0]: https://foo.bar/\n[1]: https://a.link/.\n",
		},
	}

	for idx, expectation := range expectations {
		output, err := NewConverter(strings.NewReader(expectation.input)).Convert()
		if err != nil {
			t.Errorf("error: %s", err)
		}
		if output != expectation.expectedOutput {
			t.Errorf("[%d] Failed conversion.\nInput:\n%s\nExpected:\n%s\nGot:\n%s", idx, expectation.input, expectation.expectedOutput, output)
			return
		}
	}
}

func TestExtractCode(t *testing.T) {
	expectations := []struct {
		input string
		expectedOutput string
	}{
		{
			"{{inline code}}\n",
			"`inline code`\n",
		},
		{
			"\\{\\{inline code}}\n",
			"`inline code`\n",
		},
		{
			"{code}\nblock code\n{code}\n",
			"```\nblock code\n```\n",
		},
		{
			"{code:java}\nblock code\n{code}\n",
			"```java\nblock code\n```\n",
		},

		{
			"{code:java}\nblock code {\n{code}\n",
			"```java\nblock code {\n```\n",
		},
	}

	for idx, expectation := range expectations {
		output, err := NewConverter(strings.NewReader(expectation.input)).Convert()
		if err != nil {
			t.Errorf("error: %s", err)
		}
		if output != expectation.expectedOutput {
			t.Errorf("[%d] Failed conversion.\nInput:\n%s\nExpected:\n%s\nGot:\n%s", idx, expectation.input, expectation.expectedOutput, output)
			return
		}
	}
}

func TestExtractNoFormat(t *testing.T) {
	expectations := []struct {
		input string
		expectedOutput string
	}{
		{
			"{noformat}some stuff here{noformat}\n",
			"```\nsome stuff here\n```\n",
		},
	}

	for idx, expectation := range expectations {
		output, err := NewConverter(strings.NewReader(expectation.input)).Convert()
		if err != nil {
			t.Errorf("error: %s", err)
		}
		if output != expectation.expectedOutput {
			t.Errorf("[%d] Failed conversion.\nInput:\n%s\nExpected:\n%s\nGot:\n%s", idx, expectation.input, expectation.expectedOutput, output)
			return
		}
	}
}
