from ipdb import set_trace
import time, os, threading, json
from typing import List, Dict
from dataclasses import dataclass
from openai import OpenAI

client = OpenAI(
    # This is the default and can be omitted
    # api_key=os.environ.get("OPENAI_API_KEY"),
    api_key="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"
)

# DEFAULT_MODEL = "gpt-4o"
DEFAULT_MODEL = "gpt-4.5-preview"

class Node:
    def __init__(self, id, title, question, parent=None, is_instruction=False, model=DEFAULT_MODEL, **params):
        self.id = id
        self.title = title
        self.model = model
        self.question = question
        self.is_instruction = is_instruction
        self.response_content = ""
        self.response_id = None
        self.last_answered = -1
        self.time_taken = 0

        # All parent Nodes that can lead to this via a question
        # Key is the parent ID and the value is a note for the question
        self.parent = parent
        self.parms = params

    def to_json(self):
        return {
            "model": self.model,
            "title": self.title,
            "question": self.question,
            "parentid": self.parent.id if self.parent else None,
            "is_instruction": self.is_instruction,
            "last_answered": self.last_answered,
            "time_taken": self.time_taken,
            "id": self.id,
            "response": self.response_content
        }

    @property
    def root(self):
        if not self.parent:
            return self
        return self.parent.root

def callnode(node, t=-1, callback=None):
    if t < 0: t = time.time()

    # Make sure parents are all called first
    if node.parent and node.parent.last_answered < t and not node.parent.is_instruction:
        callnode(node.parent, t, callback)

    # Now call us
    params = dict(
        # model="gpt-4o",
        model=node.model,
        temperature=0.7,
        input=node.question,
    )
    print(f"=" * 80)
    print("Asking Question: ")
    print(node.question)
    start_time = time.time()
    if node.parent:
        if node.parent.is_instruction:
            params["instructions"] = node.parent.question
            print("Using Instructions: ", node.parent.question[:100], "...")
        else:
            params["previous_response_id"] = node.parent.response.id
    node.response = client.responses.create(**params)
    node.response_content = node.response.output_text
    node.time_taken = time.time() - start_time
    node.last_answered = time.time()
    print(f"-" * 80)
    print(f"Got Response: ")
    print(node.response.output_text)
    if callback: callback(node)
    return node.response

@dataclass
class Project:
    company: str
    name: str
    timeline: str
    description: str = ""

@dataclass
class Question:
    rootfolder: str
    area: str
    subtopic: str
    content: str
    id: str
    tags: List[str] = None
    projects: List[Project] = None
    leafnodes: Dict[str, Node] = None
    lock: threading.Lock = None
    threads = {}

    @property
    def base_id(self):
        area = "_".join([p.strip() for p in self.area.split(" ") if p.strip()])
        subtopic = "_".join([p.strip() for p in self.subtopic.split(" ") if p.strip()])
        return f"{area}_____{subtopic}_____{self.id}"

    def to_json(self):
        return {
            "qinfo": {
                "area": self.area,
                "subtopic": self.subtopic,
                "id": self.id,
                "tags": self.tags or [],
                "content": self.content,
            }, 
        }

    def add_project(self, proj: Project, leaf: Node):
        self.lock = self.lock or threading.Lock()
        self.projects = self.projects or []
        self.projects.append(proj)

        self.leafnodes = self.leafnodes or {}
        self.leafnodes[proj.name] = leaf

        self.folder_for_project(proj, True)

    def folder_for_project(self, project, ensure=False):
        if type(project) is int: project = self.projects[project]
        area = "_".join([p.strip() for p in self.area.split(" ") if p.strip()])
        subtopic = "_".join([p.strip() for p in self.subtopic.split(" ") if p.strip()])

        pname = project.name
        pname = "".join([p.strip() for p in pname.split(" ") if p.strip()])
        folder = os.path.join(self.rootfolder, f"{area}_____{subtopic}_____{self.id}", pname)
        # Ensure this folder exists
        if not os.path.isdir(folder):
            print("Folder does not exist: ", folder)
            if ensure: os.makedirs(folder, exist_ok=True)
        return folder

    def path_for_project_file(self, project, name, extension):
        if type(project) is int: project = self.projects[project]
        fname = name if not extension else f"{name}.{extension}"
        return os.path.join(self.folder_for_project(project), fname)

    def answer_for_all_projects(self):
        for proj in self.projects:
            self.answer_for_project(proj)

    def answer_for_project(self, project):
        if type(project) is int: project = self.projects[project]
        """ Starts the llm chain for a particular project. """
        answerid = self.base_id + "/" + project.name
        leaf = self.leafnodes[project.name]

        with self.lock:
            if answerid in self.threads:
                print("Already answering: ", answerid)
                return

            thread = threading.Thread(target = self._answer_for_project, args=(project,))
            self.threads[answerid] = thread
            thread.start()
        return answerid

    def _answer_for_project(self, project):
        if type(project) is int: project = self.projects[project]
        leaf = self.leafnodes[project.name]

        def oncomplete(node):
            self.save_node_content(project, node, save=True)
            if True or node is leaf:
                # Done so we can create index.md
                self.save_answer_indexmd(project)

        callnode(leaf, callback=oncomplete)
        self._answer_finished(project)

    def _answer_finished(self, project):
        if type(project) is int: project = self.projects[project]
        leaf = self.leafnodes[project.name]
        answerid = self.base_id + "/" + project.name
        with self.lock:
            if answerid in self.threads:
                del self.threads[answerid]

        # Write the necessar files
        curr = leaf
        while curr:
            fname = self.path_for_project_file(project, curr.title, "md")
            with open(fname, "w") as outfile:
                outfile.write(curr.response_content)
            curr = curr.parent
        self.save_answer_meta(project)

    def save_answer_indexmd(self, project):
        if type(project) is int: project = self.projects[project]
        fname = self.path_for_project_file(project, "index", "md")
        node = self.leafnodes[project.name]
        out = ""
        curr = node
        while curr:
            if curr.title:
                out = "# " + curr.title + "\n" + curr.response_content + "\n\n" + out
            curr = curr.parent
        with open(fname, "w") as outfile:
            outfile.write("# " + self.content + "\n\n" + out)

    def save_node_content(self, project, node, save=False):
        nodefname = self.path_for_project_file(project, node.title, "md")
        nodeinfo = node.to_json()
        nodeinfo["contentfile"]  = nodefname
        nodeinfo["response"] = ""
        if save:
            with open(nodefname, "w") as outfile:
                outfile.write(node.response_content)
        return nodeinfo

    def save_answer_meta(self, project):
        if type(project) is int: project = self.projects[project]
        leaf = self.leafnodes[project.name]
        info = self.to_json()
        info["project"] = project.name
        info["nodes"] = []
        curr = leaf
        while curr:
            nodeinfo = self.save_node_content(project, curr)
            info["nodes"].append(nodeinfo)
            curr = curr.parent

        fname = self.path_for_project_file(project, "info", "json")
        print("Writing answer metadata to: ", fname)
        with open(fname, "w") as outfile:
            outfile.write(json.dumps(info, indent=2))

# The way we save qestion is:
# rootfolder/area___subtopic___qid/<nodes.
