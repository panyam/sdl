
def prompt_for_generating_questions_one_at_a_time(module):
    return f"""
I have an interview with {module.company} for the {module.name} module

Here are the expectations:

---
{module.expectations}
---

I want you to ask me one question at a time and assess my responses for Google's L7 engineering manager rubriks along with the Meta's principles.   For each answer you will critically probe me for key aspects that will be needed for a question.   Some areas area:

1. Clear context and examples around parts that may be qualitative.
2. Motivations and very brief backstory on the answer and others.

Once you have assessed my response and identified areas of improvements please do one of the following three things:

1. Pick a previous question and its response, state/refer to the question and ask me one follow up question.
2. Or ask me a new question in another area along the interview topics I have presented above.

"""

def prompt_for_generating_n_questions(module, n = 20):
    return f"""
I have an interview with {module.company} for the {module.name} module

Here are the expectations:

---
{module.expectations}
---

I need help preparing for this information.   Let us think on how to be effective for this.  I want the top {n} questions that will be asked by {module.company} that cover the above areas by the highest likelihood.  Group and tag the questions across the above areas for easy recall.
"""

def prompt_for_answering_question(module, project, question, answerformat=answerformats.GPSF, rubric=rubriks.GOOGLE):
    return f"""
I have an interview with {module.company} for the {module.name} module

Here are the expectations:

---
{module.expectations}
---

Answer this question: `{question}`

Use the following format when answering:

{answerformat}

Ensure that you are addressing relant Meta values and principles and answering at an L7 engineering manager level and
make sure to address Meta's values and principles.

Use the context of the {project.name} project.

The timeline of the project is:

---
{project.timeline}
---
"""

def prompt_for_reviewing_answer(module, question, answer, rubric=rubriks.GOOGLE):
    return f"""
I have an interview with {module.company} for the {module.name} module

Here are the expectations:

---
{module.expectations}
---

Here is an interview question: `{question}`

Assess the following answer:

{answer}
"""
