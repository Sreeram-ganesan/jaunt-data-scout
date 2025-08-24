# Orchestration System Design

Live preview (PlantUML public server):

![System design](https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/Sreeram-ganesan/jaunt-data-scout/main/epics/orchestration-step-fns/documentation/epic-system-design.pu)

Direct links:
- SVG: https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/Sreeram-ganesan/jaunt-data-scout/main/epics/orchestration-step-fns/documentation/epic-system-design.pu&fmt=svg
- PNG: https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/Sreeram-ganesan/jaunt-data-scout/main/epics/orchestration-step-fns/documentation/epic-system-design.pu&fmt=png

Alternative: render with Kroki on the command line
- SVG:
  ```bash
  curl -sS --data-binary @epics/orchestration-step-fns/documentation/epic-system-design.pu https://kroki.io/plantuml/svg > epics/orchestration-step-fns/documentation/epic-system-design.svg
  ```
- PNG:
  ```bash
  curl -sS --data-binary @epics/orchestration-step-fns/documentation/epic-system-design.pu https://kroki.io/plantuml/png > epics/orchestration-step-fns/documentation/epic-system-design.png
  ```

Local preview options
- VS Code: install the “PlantUML” extension and open `epic-system-design.pu`.
- Docker (PlantUML server):
  ```bash
  docker run --rm -p 8080:8080 plantuml/plantuml-server:jetty
  # then browse to: http://localhost:8080/svg/ <paste diagram text>
  ```

Notes
- The PlantUML server links above fetch the diagram from the repository’s raw file on the `main` branch. If the branch name changes, update the links accordingly.
- Diagram source file: `epics/orchestration-step-fns/documentation/epic-system-design.pu`.