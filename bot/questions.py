from typing import List
from core import Node, Project, Question

ALL = {
    "Project Direction": {
        "Setting and Driving Metrics": [
            "How did you identified success metrics for a complex project? How did you choose these metrics, and how did you ensure alignment with organizational goals?",
            "How have you influenced stakeholders to adopt metrics they initially disagreed with or did not prioritize?",
        ],
        "Influencing Stakeholders": [
            "How did you influenced stakeholders to implement your solution over other strong alternatives.  What methods or frameworks did you use?",
            "How do you typically handle situations where stakeholders have conflicting priorities?",
        ],
        "Strategic Alignment": [
            "Describe your approach to ensuring projects align closely with the broader organizational objectives. Can you share a recent example?",
            "What frameworks or processes do you use to communicate project alignment clearly to senior executives?",
        ],
        "Milestones and Strategic Planning": [
            "How do you set milestones for large projects to align with strategic organizational objectives? Could you walk me through an example of your involvement in strategic planning?",
            "When initiating new projects, what tools or data sources do you typically leverage to ensure strategic alignment?",
        ],
    },
    "Project Execution": {
        "Identifying and Addressing Gaps": [
            "Share an example of a project where you identified significant execution gaps. How did you diagnose these issues, and what actions did you take to resolve them?",
        ],
        "Tracking and Ensuring Quality": [
            "Can you discuss a system or process you have built or utilized that significantly improved the tracking and execution health of complex projects?",
            "How do you scale your execution processes effectively as projects and teams grow?",
        ],
        "Empowering Teams": [
            "Describe how you empower your Engineering Managers and Technical Leads to collaborate effectively with cross-functional partners.",
            "Tell me about a time you had to course-correct a team during execution. How did you approach it?",
        ],
    },
    "Stakeholder Influence and Engagement": {
        "Influencing and Negotiation": [
            "Provide an example where you successfully negotiated a critical decision with a peer or stakeholder. What strategies were most effective?",
        ],
        "Building Strong Partnerships": [
            "How have you proactively built strong, lasting relationships across your organization? Can you share a notable partnership you developed?",
        ],
        "Managing Disagreements": [
            "Describe a situation where you faced significant disagreement with a stakeholder. How did you resolve the conflict constructively and maintain the relationship?",
        ],
    },
    "Partners": {
        "Cross-Functional Collaboration": [
            "Tell me about a particularly successful cross-functional project you led. What made it successful, and how did you ensure alignment across multiple teams?",
            ],
        "Communication of Vision": [
            "How do you typically communicate your vision to cross-functional partners? Describe any frameworks or communication methods you regularly employ."
        ],
        "Planning and Alignment": [
            "Walk me through your annual planning process. How do you ensure cross-departmental alignment (e.g., business, compliance, engineering)?",
        ],
        "Technical Collaboration Example": [
            "Describe how you would approach designing a client-server API for a rich document editor. What considerations would you prioritize to ensure successful integration across teams?",
        ]
    },
    "Resolving Conflict": {
        "Conflict Resolution Examples": [
            "Can you share an example where you resolved a challenging disagreement within your team? What frameworks or approaches did you use?",
        ],
        "Empathy and Understanding": [
            "Describe a time you deeply disagreed with someone's perspective. How did you demonstrate empathy and still move the project forward?",
        ],
        "Proactive Conflict Prevention": [
            "What systems or processes have you established in your teams to proactively address and resolve conflicts?"
        ]
    },
    "Growing Continuously": {
        "Learning from Failure": [
            "Discuss a project you were passionate about that ultimately failed. What key insights did you gain, and how did you translate this into improved practices?",
        ],
        "Feedback and Growth": [
            "How do you actively seek and integrate constructive criticism in your leadership style? Provide an example where feedback notably improved your approach or project outcomes.",
        ],
        "Scaling Self-Awareness": [
            "What practices or rituals do you encourage your teams to adopt to continuously grow and increase self-awareness (e.g., retrospectives, SWOT analyses)?"
        ]
    },
    "Embracing Ambiguity": {
        "Operating Under Uncertainty": [
            "Tell me about a situation where you had to make critical decisions with incomplete information. What strategies did you use to manage uncertainty?",
        ],
        "Pivoting Teams": [
            "Describe an instance where you had to pivot your team abruptly due to shifting organizational priorities. How did you handle the change to keep your team productive and focused?",
        ]
    },
    "Communicating Effectively": {
        "Effective Communication Practices": [
            "Describe your approach to effectively communicating complex technical information to various audiences, including your teams, cross-functional partners, and senior executives.",
        ],
		"Identifying and Addressing Communication Gaps": [
            "Provide an example where you identified significant communication gaps within or between teams. How did you address these gaps and ensure clarity and alignment?",
        ],

		"Coaching for Communication": [
            "How do you coach your team members to communicate more effectively with their stakeholders? Share a specific scenario where your coaching had a positive impact."
        ],
    },
}
