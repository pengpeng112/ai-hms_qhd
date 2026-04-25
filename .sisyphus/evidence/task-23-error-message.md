T23 Evidence - Error Message Improvements
==========================================

Status: PASS (dialysis / DictConfig / patient error handling verified)

Verification:
1. restClient.ts contains:
   - getErrorMessage(): extracts backend error message or returns generic error
   - getRequestErrorKind(): categorizes errors as auth/forbidden/not_found/network/server_error
   - getTreatmentLoadErrorMessage(): provides specific messages for treatment loading failures

2. DialysisExecution.tsx uses:
   - getErrorMessage(error) for general API failures
   - getTreatmentLoadErrorMessage(error) for treatment-specific failures
   - getRequestErrorKind(error) for error-state separation (T19)

3. All child tabs (PreAssessment, PostAssessment, MidMonitoring, TodayPrescription, Verification, MedicalOrders) use getErrorMessage consistently.

4. Error categories distinguished:
   - auth/forbidden -> authentication error
   - not_found -> missing treatment record (different from empty)
   - network -> network connectivity issue
   - server_error -> server-side failure

5. No fixed/placeholder API error messages remain in treatment flow.

6. DictConfig / PatientList / PatientDetail / patient-detail tabs now route async failures through getErrorMessage().

7. Remaining literal message.error() calls are intentional local validation prompts (required fields / selection guards), not API failures.

8. Local validation:
   - backend: go test ./... ✅
   - frontend: npm run build ✅ (only existing chunk-size warning)

T23 Status: COMPLETE - Error messages are improved and consistent across the scoped pages.
