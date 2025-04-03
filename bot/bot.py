
import os, threading
import answerformats, rubriks
from core import Node, Question, Project, callnode

# Data specific to the project
project_data = {
    "item1": "content1",
    "item2": "content2",
    "item3": "content3",
}

def combined_round():
    return """
My next interview is with Meta on the Project Retrospective and Behavioral Round.   It will be in 8 areas:

• Project direction:
    -   How do you set and drive success metrics?
    -   How do you influence stakeholders to implement the most impactful solution?
    -   Discuss how you align your projects to the objectives of the organization at large.
    -   How do you establish milestones and drive strategic planning for your orgnaization goals (talk about
        participating in strategic planning)?
    -   Talk about bringing about stakeholder alignment on projects you have initiated or driven.
    -   When you initiate projects how do you investigate the particular problem to gauge alignment with org goals?
        Talk about customer data, operational reports, and strategic business drivers.
    -   How would you quantify success metrics and explain specifically how you would chose the metrics/baselines you nominated?
    -   How you would gain buy-in from critical stakeholders? How you would align conflicting priorities among stakeholders to drive consensus.
    -   What techniques/frameworks/process do you have for communicating the strategic importance of projects to senior leadership or executives?

• Project execution: How do you identify and address gaps in execution?  What tools and/or systems have you utilized or built to ensure quality work across complex problems?  How do you track execution health and statuses?  What proceses have you put in place to scale processes?  How do you empower your teams, EMs and TLs to execute and collaborate with partners?

• Stakeholder influence and engagement: When have you had to influence and negotiate with stakeholders and peers? How do you build strong partnerships across the organization? How do you nurture relationships, foster engagement and empower your stakeholders?  How do you handle disagreements and address and assess key issues with your partners?

• Partners: How do you partner with cross-functional partners and stakeholders? What are some successful collaborations that you’ve had? How do you communicate your vision to your partners (talk about your framework and manifesto).  What is your plannning process through out the year?  How do you align variou sorgs like business, compliance etc?  How do you build aliances?  How do you empower your teams and managers to work with your parters?  How do you get everybody on the same page?  Tell me how you’d design a client-server API to build a rich document editor

• Resolves Conflict: What kind of disagreements have you had with colleagues and/or team members? How have you resolved them? Can you empathize with people whose points of view differ radically from yours?  When resolving conflicts what kind of frameworks and processes do you setup for your teams (and for yourself) for the future?   How do you fix existing processes?

• Grows Continuously: Discuss a project that you led and were really passionate about that failed. Why did it fail and what did you learn? How did take constructive criticism as an opportunity to improve?  What best practices did you put in place as a result of this or did you improve as a result?  Talk about Receiving feedback, retrospectives, scaling self awareness for you and your teams, SWOT.

• Embraces Ambiguity: How do you operate in an ambiguous and quickly changing environment? Can you make quality decisions and sustain productivity when missing information? How did you react when you had to pivot your team away from a project due to a shift in priority?

• Communicates Effectively: How well do you communicate with teams and cross-functional partners?  How do you coach your team to effectively communicate with their stakeholders?   How do you study the audience?  Identify gaps in communication styles, address the gaps, help teams resonate with your message and persuade all audience types?   How do you measure this?   How do you deliver complex pieces of information to teams, partners, leadership etc?
"""

def instructions_for_question(question, project):
    prompt = f"""
{combined_round()}
"""
    if project:
        prompt += f"""
Make sure in your answers you use 'I' instead of 'We' whereever possible.  Key is to show your actions and impact.  When answering the question pick examples from the {project.name} project.  Its timeline is:

{project.timeline}
"""
    prompt += "Answer this question: `{question.content}`"
    return Node("root", "", question=prompt, is_instruction=True)

def instructions_for_question2(question, project, structure=answerformats.GPSF):
    prompt = f"{combined_round()}"
    prompt += f"\n\nAnswer this question: `{question.content}`\n\nUse the following format when answering:"
    prompt += f"\n---\n{structure}\n---"
    prompt += f"\n\nEnsure that you are addressing relevant Meta values and principles and answering at an L7 engineering manager level and make sure to address Meta's values and principles."
    if project:
        prompt += f"\n\nMake sure in your answers you use 'I' instead of 'We' whereever possible.  Key is to show your actions and impact.  When answering the question pick examples from the {project.name} project.  Its timeline and details are:"
        prompt += f"\n---\n{project.timeline}\n---"
    return Node("root", "OneShotAnswer", question=prompt)

def nodes_for_chained_answer(question, project):
    root = instructions_for_question(question, project)
    node0 = root # Node("briefsummary", "Brief Summary", question=f"First briefly summarize this particular question (only) to highlight the real intent behind it in less than 15 words.", parent=root)
    node1 = Node("oneminute", "One minute answer", question=f"Now give a 1 minute version that captures the situation, context, your specific actions/frameworks and results.", parent=node0)
    node2 = Node("twominute", "Two Minute Answer", question=f"""
        Great now come up with a 2 minute version that provides a timeline view of all the triggers and the actions you took to assess and address the situations along with who was involved in this action (teams, TLs, EMs, partners and leaders etc).
    """, parent=node1)

    node3 = Node("detailed", "Detailed Response", question=f"""
        Great now give me a detailed five minute version that dives deep into actions you took across various aspects of the problem solving including partners, team members, leadership and customers (showed in the 2 minute version).  This should also include frameworks you had in place, what new frameworks/processes you put in place and evolved existing ones to empower teams and partners in scaling their impact.  Make sure to assess this for any qualitative aspects and then fill in the details so that there is sufficient depth to address any qualitative statements.
    """, parent=node2, model = "gpt-4.5-preview")

    node4 = Node("flowchart", "Summary Flowchart", question="Great now can you also give me a summary flowchart diagram linking all the actions you took?", parent = node3, model = "gpt-4.5-preview")
    return node4

def nodes_for_answer_with_review(question, project):
    root = instructions_for_question2(question, project)
    node1 = Node("reviewed", "Reviewed Answer", question=f"""Great now I want you to play the role of the reviewer and review this answer at a senior engineering manager level.  I want you to review the answer critically and address areas where there is insufficient and fix and rewrite the candidate's answer by addressing all the areas of improvement.""", parent=root)
    return node1

class Session:
    """ A session describes a top level question to be run on all projects along with its nodes"""
    def __init__(self, output_folder="./responses"):
        self.output_folder = output_folder
        self.questions = []

    def load_questions(self, qdict=None, projs=None):
        qdict = qdict or {}
        if not qdict:
            import questions
            qdict = questions.ALL
        if not projs:
            import projects
            projs = [projects.MemoryAutoScaler, projects.JobSchedulerService, projects.CostTrackingSystem]
        for area, subtopic in qdict.items():
            for topicname, qlist in subtopic.items():
                for index, question in enumerate(qlist):
                    q = Question(rootfolder=self.output_folder, id=f"{index}", area=area,subtopic=topicname,content=question)
                    self.questions.append(q)
                    for proj in projs:
                        leaf = nodes_for_answer_with_review(q, proj)
                        q.add_project(proj, leaf)
        return self

def generate_html(session, outfolder):
    pass
