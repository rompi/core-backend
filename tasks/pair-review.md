# 🧩 Pair Review Instruction — Analysis & Plan

## 🎯 Purpose  
This review ensures that the **analysis and plan** (from the issue-specific docs the user references) are logical, complete, and aligned with objectives.  
Your goal as a reviewer is to **clarify**, **validate reasoning**, and **highlight any unclear or inconsistent parts**.

---

## 🪜 Steps

### 1. Read the Markdown File
- Identify the relevant issue number from the user input (e.g., `issue-123`).
- Open the plan files for that issue (typically `docs/plans/issue-<NUMBER>/analysis.md` and related Markdown files).
- Review all sections (context, analysis, plan, metrics, timeline, etc.).
- Understand the intent behind each section before commenting.

### 2. Check for Clarity
Ask yourself:
- Does each point make sense without extra explanation?  
- Are there any vague, ambiguous, or missing details?  
- Do assumptions or dependencies need to be clarified?  

If something is **unclear**, write it under the “Clarification Needed” section.

### 3. Check for Logic & Alignment
- Are the conclusions logically derived from the analysis?  
- Are the proposed plans realistic and actionable?  
- Do timelines and goals align with the stated objectives or context?  
- Are dependencies or risks clearly mentioned?

If something feels **inconsistent or illogical**, write it under “Not Making Sense”.

### 4. Check Completeness
- Does the document cover all necessary aspects (problem, analysis, action plan, expected outcomes)?  
- Are key stakeholders, data sources, or success metrics missing?  
- Are follow-up or next steps clear?

---

## 🗒️ Output Format

When you finish, provide your review using this structure:

### ✅ Summary
A brief 2–3 sentence summary of your understanding of the analysis and plan.

### 💬 Clarification Needed
List all unclear or incomplete points.  
Example:
- [ ] Please clarify what “data sync layer” refers to in this context.  
- [ ] Timeline for Phase 2 is missing exact dates.

### ⚠️ Not Making Sense / Needs Revision
List anything illogical or inconsistent.  
Example:
- [ ] Action 3 assumes integration with Tool X, but earlier context says we haven’t built that yet.  
- [ ] The success metric doesn’t match the stated problem.

### 💡 Optional Suggestions
(Optional) Add ideas for improvement, rewording, or simplification.

---

## 🧭 Reviewer’s Role
Your job is **not to rewrite**, but to:
- Ensure clarity and logical flow.
- Ask questions where reasoning or context is unclear.
- Help the author see potential blind spots.

The **final decision** on changes remains with the document’s author.
