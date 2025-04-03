from dataclasses import dataclass
from core import Project

MemoryAutoScaler = Project(company="Google", name = "Memory Autoscaler",
                           timeline="""
Project Overview: Memory Autoscaler for Google Cloud Dataflow
Total Duration: 11 months (Planning through full rollout)

Primary Goal:
Build a predictive, adaptive, and automated memory optimization system for Google Cloud Dataflow jobs, preventing memory-related failures (OOM errors) and eliminating excessive memory allocations, thus significantly reducing operational costs.

1. Planning & Scoping (Month 1-2)
Activities:

Gathered extensive customer feedback and Dataflow job failure metrics.

Analyzed historical execution patterns (terabytes of data).

Established clear objectives: dynamic memory management, cost optimization, predictive analytics.

Key Stakeholders Identified:

Dataflow Product Management

Compute Engine & Regional MIGs

Cloud Billing and Observability teams

BigQuery and AI/ML teams

Aligned Google Principle:

"Focus on long-term impact" (building predictive ML capabilities that scale and provide sustained savings)

"Move fast" (urgent identification of customer pain points and immediate scoping)

Aligned Meta Principle:

"Focus on long-term impact" (investment in predictive models providing sustained benefits to customers)

"Move fast" (rapid identification of critical pain points and immediate clarity on project objectives)

2. Cross-functional Alignment & Resource Allocation (Month 3)
Activities:

Created a cross-team alignment strategy, defining roles, responsibilities, and deliverables.

Negotiated resource allocation from Compute, Billing, and AI teams.

Established communication channels and defined success metrics.

Success Metrics Defined Clearly:

Reduce OOM job failures by 30%

Cut memory allocation costs by 20%

Achieve autoscaling latency of < 1 minute for real-time adjustments

Aligned Google Principle:

"Be direct and respect your colleagues" (clear resource negotiation and open feedback channels)

Aligned Meta Principle:

"Meta, Metamates, me" (shared responsibility, collective success via clear team roles and accountability)

3. Architecture & ML Model Development (Months 4-6)
Activities:

Designed architecture integrating Compute Engine, Cloud Billing, BigQuery, and ML platforms.

Developed predictive models leveraging historical Dataflow job execution data.

Extensive validation and fine-tuning of models for accuracy.

Technical Complexity:

ML-driven autoscaling (vertical and horizontal)

Integration across multiple GCP services (Compute Engine, AI platform, BigQuery)

Aligned Google Principle:

"Build awesome things" (delivering advanced ML-based prediction for unprecedented memory optimization at cloud scale)

Aligned Meta Principle:

"Build awesome things" (developing awe-inspiring capabilities through ML innovation)

4. Development, Integration, and Initial Testing (Months 7-9)
Activities:

Developed autoscaler engine integrated with Compute Engine MIGs for automated memory tuning.

Integrated predictive ML models into autoscaler control plane.

Extensive internal testing, validation, and iterative improvement.

Tools & Systems Utilized:

BigQuery ML, Google Cloud AI Platform, Compute Engine, Cloud Storage (for checkpointing), Cloud Monitoring, and IAM (security and compliance)

Aligned Google Principle:

"Move fast" (rapid iteration, continuous internal releases for fast feedback)

Aligned Meta Principle:

"Move fast" (accelerated testing and iterative improvement)

5. Beta Release & Customer Feedback (Month 10)
Activities:

Limited beta rollout with key enterprise customers.

Collected real-world usage data, monitored performance, cost savings, and customer satisfaction.

Responded to customer feedback quickly, refining models, algorithms, and user experiences.

Customer Impact:

Immediate positive feedback on cost efficiency and reliability improvements.

Fine-tuned predictive accuracy from real customer data.

Aligned Google Principle:

"Serve Everyone" (immediate customer-centric iterations based on feedback)

Aligned Meta Principle:

"Give people a voice" (actively integrating customer feedback to refine the product)

"Keep People Safe and Protect Privacy" (strict compliance with privacy and security requirements via IAM)

6. Full-scale Launch, Monitoring, and Retrospective (Month 11)
Activities:

Global rollout to all Google Cloud Dataflow customers.

Comprehensive monitoring of success metrics: surpassed original goals—OOM errors reduced by 40%, memory costs reduced by 30%.

Project retrospective identifying wins, improvements, and future learnings.

Results & Organizational Impact:

Millions in customer savings achieved

Set a new standard for AI-driven resource optimization at scale

Reinforced Google's leadership in cloud-based cost-efficient AI services

Aligned Google Principle:

"Live in the future" (establishing new benchmarks and future standards in AI-driven cloud optimization)

Aligned Meta Principle:

"Live in the future" (driving innovation forward and setting a new industry standard)

Summary of Scale & Magnitude:
Scale: Global rollout impacting millions of Dataflow jobs per day, processing petabytes of data, real-time ML-driven optimization at unprecedented cloud scale.

Magnitude: 40% reduction in OOM errors, 30% reduction in unnecessary costs, comprehensive collaboration across multiple large-scale Google infrastructure teams.

This project’s successful alignment to both Google's and Meta’s principles reflects a deep commitment to long-term customer value, innovation through AI-driven solutions, rapid iteration, direct communication, and customer-centric feedback loops.
""")

JobSchedulerService = Project(company="Google", name = "Job Scheduler Service",
                           timeline="""
## 🔍 **Execution Plan Overview (10 Months):**

Here's the month-by-month breakdown grouped into logical phases:

### **Phase 1: Planning & Stakeholder Alignment (Month 1-2)**

-   **Goals:**
    -   Define clear project scope and success criteria.
    -   Identify and engage key stakeholders.
    -   Align on business objectives and technical approach.
        
-   **Key Activities:**
    -   Conduct stakeholder workshops across teams (**Compute Engine, Spanner, Cloud Storage, IAM, Observability**).
    -   Perform detailed market analysis (user stories, use-cases).
    -   Finalize functional and non-functional requirements, e.g., **low latency (<1s scheduling), DAG-based scheduling**.
        
-   **Deliverables:**
    -   Formal PRD (Product Requirements Document).
    -   High-level architectural diagrams.
        
-   **Alignment:**
    -   Google: **Project scoping and organization, PA-wide business objectives**
    -   Meta: **Be direct and respect your colleagues, Focus on long-term impact**
        
### **Phase 2: Technical Architecture & Detailed Design (Month 3-4)**

-   **Goals:**
    -   Develop robust architecture ensuring scalability, availability, and regional failover.
    -   Identify and resolve technical trade-offs.
        
-   **Key Activities:**
    -   Detailed architecture reviews with **Spanner DB, Pub/Sub messaging, Compute Engine (VM provisioning), and Cloud Storage** for checkpoints and state management.
    -   Selection of technology stack and protocols (**REST & gRPC interfaces**).
    -   Design observability mechanisms (**Cloud Logging, Cloud Trace, Monitoring**).
        
-   **Deliverables:**
    -   Detailed design documents reviewed and approved by stakeholders.
    -   Proof-of-concept prototypes demonstrating feasibility (initial DAG scheduler).
        
-   **Alignment:**
    -   Google: **Managing projects and execution at PA-level objectives, Navigating project ambiguity**
    -   Meta: **Build awesome things, Live in the future** (anticipating future scaling)

### **Phase 3: Development & Iterative Execution (Month 5-8)**

-   **Goals:**
    -   Deliver MVP (Minimum Viable Product) quickly for early customer validation.
    -   Iteratively develop advanced functionalities like event-driven execution.
        
-   **Key Activities:**
    -   Developed core components:
        -   **Job Scheduler** (Cron & DAG scheduling).
        -   **Job Runner** (VM provisioning, Memory Autoscaler integration)
        -   **Event Listener** (Pub/Sub integration for real-time triggers).
    -   Bi-weekly agile sprints, rigorous internal demos for rapid iteration.
    -   Close collaboration with cross-functional teams to handle integration complexities and dependencies.
        
-   **Deliverables:**
    -   Working MVP for early internal and select-customer testing.
    -   Incremental feature rollouts every 2-3 weeks.
        
-   **Alignment:**
    -   Google: **Organizational and career development (cross-team growth, dynamic execution)**
    -   Meta: **Move fast, Be direct and respectful, Build connection and community** (via frequent and transparent demos and feedback loops)
        
### **Phase 4: Testing, Validation, and Launch Prep (Month 9)**

-   **Goals:**
    -   Ensure high-quality release candidate (RC).
    -   Conduct comprehensive testing (load, integration, security, usability).
        
-   **Key Activities:**
    -   Performed extensive scalability and stress tests (1000s concurrent jobs).
    -   Security audit and compliance reviews (**IAM integration**).
    -   Customer beta testing, incorporating direct feedback.

-   **Deliverables:**
    -   Finalized RC version.
    -   Published documentation, training materials, and customer case studies.
        
-   **Alignment:**
    -   Google: **High availability, reliability, and strong consistency via Spanner**
    -   Meta: **Keep people safe and protect privacy, Promote economic opportunity**
        
### **Phase 5: Launch, Monitoring & Iterative Enhancement (Month 10)**

-   **Goals:**
    -   Successful launch and adoption tracking.
    -   Rapid response to any post-launch issues.
        
-   **Key Activities:**
    -   Coordinated marketing, communication, and product launch announcements.
    -   Active real-time monitoring post-launch (**Observability stack**).
    -   Collected usage metrics, error rates, and user satisfaction feedback.
    -   Rapid hotfixes and incremental enhancements.
        
-   **Deliverables:**
    -   Successful launch announcement.
    -   Dashboard of adoption metrics and operational performance.
        
-   **Alignment:**
    -   Google: **Owning and driving large and complex initiatives into production, visible at executive-level objectives**
    -   Meta: **Meta, Metamates, me (Collective ownership of success), Move fast (rapid iteration post-launch)**
""")

CostTrackingSystem = Project(company="Google", name = "Cost Usage Tracking",
                           description="Purpose: To build a real-time, predictive cost-tracking system for Google Cloud Dataflow, delivering fine-grained visibility down to per-job, per-stage, per-worker levels, addressing significant gaps in Google Cloud’s existing billing visibility.",
                           timeline="""
## 🗓 **End-to-End Execution Timeline (~6 months)**

### **Month 1: Project Definition & Stakeholder Alignment**

-   **Scope Definition:**
    -   Clearly define functional & non-functional requirements (real-time analytics, predictive forecasting, customizable dashboards).

-   **Success Metrics:**
    -   Reduce customer cost overruns by ≥30%.
    -   Enable predictive insights with <1s latency.
        
-   **Key Stakeholder Identification:**
    -   Google Cloud Billing, Compute Engine, Dataflow PM, BigQuery ML, IAM teams.
        
-   **Strategic Alignment:**
    -   Ensure alignment with larger organizational objectives around operational efficiency and cloud cost transparency.
        
-   **Deliverables:**
    -   Project Charter approved by stakeholders.
    -   Clear roadmap (milestones, timelines).
        
### **Month 2: Detailed System Design & Architecture**

-   **Architecture Planning:**
    -   Finalize the High-Level Architecture (Cost Analyzer, Usage Tracker, Predictive ML, Observability, IAM).
    -   Integration design with Google Cloud APIs (Cloud Billing, Compute Engine, BigQuery ML).
        
-   **Technical Alignment (Google principles):**
    -   Architected for scale and reliability—multi-region, high availability, strong consistency (Spanner-backed).
        
-   **Meta Principles Alignment:**
    -   "Build awesome things" with an innovative predictive forecasting model, setting industry standards for cost analytics.
        
-   **Deliverables:**
    -   System design documentation.
    -   Stakeholder design review completed.

### **Month 3: MVP Development and Initial Integrations**

-   **Engineering Execution:**
    -   Stand up core data infrastructure (Spanner, Cloud Storage, Compute Engine integration).
    -   Initial ML model training pipeline established with BigQuery ML.
        
-   **Principle Alignment (Meta "Move fast"):**
    -   Rapid MVP launch to quickly validate value with internal teams.
        
-   **Collaboration:**
    -   Regular cross-team syncs (Compute Engine, Billing APIs, Dataflow product team).
        
-   **Deliverables:**
    -   MVP showcasing basic cost tracking per-job & per-worker.
        
### **Month 4: Iterative Development & ML Integration**

-   **ML Predictive Model Development:**
    -   Refine ML forecasting accuracy using historical Dataflow job data.
    -   Begin real-time integration of ML predictions with live billing streams.
        
-   **Real-time Monitoring & Observability:**
    -   Integrate Cloud Monitoring and Cloud Logging for real-time cost spike detection and alerts.
        
-   **Principle Alignment (Google):**
    -   Achieve technical excellence in ML-driven prediction reliability.
        
-   **Deliverables:**
    -   ML-predictive cost dashboard with alert capabilities.

### **Month 5: Scale Testing & Performance Optimization**

-   **Load & Stress Testing:**
    -   Rigorous testing across millions of simulated Dataflow jobs per day.
    -   Optimization of Spanner queries for sub-second latency.
        
-   **Operational Readiness:**
    -   Failover testing, multi-region deployment for high availability.
        
-   **Alignment to Google & Meta "Operational Excellence":**
    -   System optimized for massive global scale.
        
-   **Deliverables:**
    -   Verified scalability and performance benchmarks.
    -   Disaster Recovery (DR) & Business Continuity Plan (BCP).

### **Month 6: Final Launch & Stakeholder Enablement**

-   **Full Production Deployment:**: Deployment of integrated and standalone modes available to customers via Cloud Console.
-   **Training & Documentation:**: Comprehensive documentation, training sessions, and onboarding for customer support teams.
-   **Principle Alignment ("Meta, Metamates, Me" & Google's Level 7 leadership):**: Deep investment in organizational capability building, ensuring broad adoption.
-   **Deliverables:**
    -   Full public launch.
    -   Customer engagement plan executed (early adopter feedback, customer workshops).
""")
