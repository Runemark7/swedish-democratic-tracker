import { api } from "@/shared/api-client";
import type { components } from "@/shared/api-contract";

type AreaTimeSeries = components["schemas"]["AreaTimeSeries"];
type BudgetComparison = components["schemas"]["BudgetComparison"];
type BudgetYear = components["schemas"]["BudgetYear"];
type BudgetYearDetail = components["schemas"]["BudgetYearDetail"];
type ExpenditureArea = components["schemas"]["ExpenditureArea"];

export const budgetApi = {
  listYears: () => api.get<BudgetYear[]>("/budget/years"),
  getYear: (year: number) => api.get<BudgetYearDetail>(`/budget/years/${year}`),
  compareYears: (baseYear: number, compareYear: number) =>
    api.get<BudgetComparison>(`/budget/years/${baseYear}/compare/${compareYear}`),
  listAreas: () => api.get<ExpenditureArea[]>("/budget/areas"),
  getAreaTimeSeries: (code: string) => api.get<AreaTimeSeries>(`/budget/areas/${code}`),
};
