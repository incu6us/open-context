#!/bin/bash

# Example script to add Jenkins documentation
# This demonstrates how to programmatically add new documentation

set -e

LANG_DIR="data/jenkins"
TOPICS_DIR="$LANG_DIR/topics"

echo "Creating Jenkins documentation structure..."
mkdir -p "$TOPICS_DIR"

echo "Creating metadata..."
cat > "$LANG_DIR/metadata.json" << 'EOF'
{
  "name": "jenkins",
  "displayName": "Jenkins",
  "description": "Jenkins CI/CD automation and pipeline documentation"
}
EOF

echo "Adding Pipeline Basics topic..."
cat > "$TOPICS_DIR/pipeline-basics.json" << 'EOF'
{
  "id": "pipeline-basics",
  "title": "Jenkins Pipeline Basics",
  "description": "Introduction to Jenkins declarative and scripted pipelines",
  "keywords": ["pipeline", "jenkinsfile", "ci", "cd", "declarative", "scripted"],
  "content": "# Jenkins Pipeline Basics\n\nJenkins Pipeline is a suite of plugins that supports implementing and integrating continuous delivery pipelines.\n\n## Declarative Pipeline\n\n```groovy\npipeline {\n    agent any\n    \n    stages {\n        stage('Build') {\n            steps {\n                echo 'Building...'\n                sh 'make build'\n            }\n        }\n        stage('Test') {\n            steps {\n                echo 'Testing...'\n                sh 'make test'\n            }\n        }\n        stage('Deploy') {\n            steps {\n                echo 'Deploying...'\n                sh 'make deploy'\n            }\n        }\n    }\n    \n    post {\n        always {\n            echo 'Pipeline completed'\n        }\n        success {\n            echo 'Pipeline succeeded'\n        }\n        failure {\n            echo 'Pipeline failed'\n        }\n    }\n}\n```\n\n## Scripted Pipeline\n\n```groovy\nnode {\n    stage('Build') {\n        echo 'Building...'\n        sh 'make build'\n    }\n    \n    stage('Test') {\n        echo 'Testing...'\n        sh 'make test'\n    }\n    \n    stage('Deploy') {\n        echo 'Deploying...'\n        sh 'make deploy'\n    }\n}\n```\n\n## Environment Variables\n\n```groovy\npipeline {\n    agent any\n    \n    environment {\n        APP_NAME = 'my-app'\n        VERSION = '1.0.0'\n    }\n    \n    stages {\n        stage('Build') {\n            steps {\n                echo \"Building ${env.APP_NAME} version ${env.VERSION}\"\n            }\n        }\n    }\n}\n```"
}
EOF

echo "Adding Plugin Development topic..."
cat > "$TOPICS_DIR/plugin-development.json" << 'EOF'
{
  "id": "plugin-development",
  "title": "Jenkins Plugin Development",
  "description": "Guide to developing Jenkins plugins",
  "keywords": ["plugin", "development", "extension", "java", "maven"],
  "content": "# Jenkins Plugin Development\n\nJenkins plugins extend Jenkins functionality and are written in Java.\n\n## Plugin Structure\n\n```\nmy-plugin/\n├── pom.xml\n└── src/\n    └── main/\n        ├── java/\n        │   └── com/example/\n        │       └── MyPlugin.java\n        └── resources/\n            └── index.jelly\n```\n\n## Basic Plugin (pom.xml)\n\n```xml\n<project>\n    <parent>\n        <groupId>org.jenkins-ci.plugins</groupId>\n        <artifactId>plugin</artifactId>\n        <version>4.40</version>\n    </parent>\n    \n    <groupId>com.example</groupId>\n    <artifactId>my-plugin</artifactId>\n    <version>1.0-SNAPSHOT</version>\n    <packaging>hpi</packaging>\n    \n    <name>My Plugin</name>\n    <description>Example Jenkins plugin</description>\n</project>\n```\n\n## Simple Builder Plugin\n\n```java\npackage com.example;\n\nimport hudson.Extension;\nimport hudson.Launcher;\nimport hudson.model.AbstractBuild;\nimport hudson.model.AbstractProject;\nimport hudson.model.BuildListener;\nimport hudson.tasks.Builder;\nimport hudson.tasks.BuildStepDescriptor;\n\npublic class MyBuilder extends Builder {\n    @Override\n    public boolean perform(AbstractBuild<?, ?> build, \n                          Launcher launcher, \n                          BuildListener listener) {\n        listener.getLogger().println(\"Hello from My Plugin!\");\n        return true;\n    }\n    \n    @Extension\n    public static final class DescriptorImpl \n        extends BuildStepDescriptor<Builder> {\n        \n        @Override\n        public boolean isApplicable(Class<? extends AbstractProject> jobType) {\n            return true;\n        }\n        \n        @Override\n        public String getDisplayName() {\n            return \"My Custom Builder\";\n        }\n    }\n}\n```"
}
EOF

echo "Jenkins documentation added successfully!"
echo "Restart the open-context server to load the new documentation."