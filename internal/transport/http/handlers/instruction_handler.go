package handlers

import (
	"net/http"

	"example.com/taskservice/internal/usecase/instruction"
)

type InstructionHandler struct {
	usecase instruction.UseCase
}

func NewInstructionHandler(usecase instruction.UseCase) *InstructionHandler {
	return &InstructionHandler{
		usecase: usecase,
	}
}

func (h *InstructionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req instructionMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), instruction.CreateInput{
		Scenario:      req.Scenario,
		ScenarioValue: req.ScenarioValue,
		TaskID:        req.TaskID,
		SpecificDates: req.SpecificDates,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newInstructionDTO(created))
}

func (h *InstructionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newInstructionDTO(task))
}

func (h *InstructionHandler) GetByTaskID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.usecase.GetByTaskID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newInstructionDTO(task))
}

func (h *InstructionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req instructionMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, instruction.UpdateInput{
		Scenario:      req.Scenario,
		ScenarioValue: req.ScenarioValue,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newInstructionDTO(updated))
}

func (h *InstructionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

func (h *InstructionHandler) List(w http.ResponseWriter, r *http.Request) {
	instructions, err := h.usecase.List(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]instructionDTO, 0, len(instructions))
	for i := range instructions {
		response = append(response, newInstructionDTO(&instructions[i]))
	}

	writeJSON(w, http.StatusOK, response)
}
