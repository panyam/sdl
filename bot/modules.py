
from dataclasses import dataclass
from typing import List

@dataclass
class ModuleArea:
    name: str
    expectations: str

@dataclass
class Module:
    company: str
    name: str
    expectations: str
    
PROJECT_RETRO_ROUND = Module(company="Meta", name="Project Retrospective", expectations="""
These are the instructions given to me by the Meta recruiter on the Project Retrospective round.

## What can you expect?

This discussion-based interview covers project implementation details, potentially
across the full project lifecycle. The interviewer will ask you a variety of questions
related to project management situations focusing on business acumen, strategy,
and execution. You will be expected to be able to deep dive into some of these
projects. You will want to provide enough detail to explain the project at both high
and detailed levels.

## What does Meta look for?

In your Project Retrospective interview, your interviewer will assess your
performance on three (3) focus areas:

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

## How to prepare for this round?

Recall two to three different projects you’ve been responsible for and map out the various aspects of the project. Focus on how to best describe the high and detailed levels of the project.
• Be clear and concise. The interviewer won’t have any background knowledge of your example(s), so please practice providing enough context. Avoid overly complex examples that are difficult to describe or require too much context (i.e. company-specific knowledge for which your interviewer doesn’t have a frame of reference).
• Map your project retrospective back to basic management principles.
• Provide clear examples that don’t require too many details.
• Focus on your ability to zoom out and look at each project or product more holistically.
• Think about the following for each example you are preparing:
   - How did you define project scope, key stakeholders, deliverables, and success metrics for tracking?
    - Project planning, roadmapping and prioritization
    - How did you empower people on your team to execute on project milestones?
    - Overall outcomes: big wins, failures, reflection on what you might do differently next time
    - Cross functional work
    - Roadblocks and how you removed them
    - How and when you got involved
""")

BEHAVIORAL_ROUND = Module(company="Meta", name="Behavioral", expectations="""
These are the instructions given to me by the recruiter on the Behavioral Round.

What can you expect?

The behavioral interview will consist of a 45-minute session. Your interviewer will
ask you to share stories and situations that present how you, as a leader, have
navigated the complex business problems that affected the company at large.

What do we look for?

The purpose of the behavioral interview is to assess if a leadership candidate will
thrive in Meta’s fast-paced and highly unstructured environment. To that end, we
assess candidates on five signals that correlate with success at Meta:

• Partners: How do you partner with cross-functional partners and stakeholders? What are some successful collaborations that you’ve had? How do you communicate your vision to your partners (talk about your framework and manifesto).  What is your plannning process through out the year?  How do you align variou sorgs like business, compliance etc?  How do you build aliances?  How do you empower your teams and managers to work with your parters?  How do you get everybody on the same page?  Tell me how you’d design a client-server API to build a rich document editor

• Resolves Conflict: What kind of disagreements have you had with colleagues and/or team members? How have you resolved them? Can you empathize with people whose points of view differ radically from yours?  When resolving conflicts what kind of frameworks and processes do you setup for your teams (and for yourself) for the future?   How do you fix existing processes?

• Grows Continuously: Discuss a project that you led and were really passionate about that failed. Why did it fail and what did you learn? How did take constructive criticism as an opportunity to improve?  What best practices did you put in place as a result of this or did you improve as a result?  Talk about Receiving feedback, retrospectives, scaling self awareness for you and your teams, SWOT.

• Embraces Ambiguity: How do you operate in an ambiguous and quickly changing environment? Can you make quality decisions and sustain productivity when missing information? How did you react when you had to pivot your team away from a project due to a shift in priority?

• Communicates Effectively: How well do you communicate with teams and cross-functional partners?  How do you coach your team to effectively communicate with their stakeholders?   How do you study the audience?  Identify gaps in communication styles, address the gaps, help teams resonate with your message and persuade all audience types?   How do you measure this?   How do you deliver complex pieces of information to teams, partners, leadership etc?

We may ask you to:
• Discuss and share examples of how you deal with conflict.
• Talk about how and why you become a people manager.
• Describe a few of your peers at your company and the type of relationship
you have with them.
• Discuss available details of past and current projects (both successes and
failures)

How to prep?

Just like with other aspects of the interview, it is important to prepare ahead of
time for interviews that are designed to get to know your background better. In this
interview, you should focus on teamwork, leadership, and mentorship qualities.
""")
