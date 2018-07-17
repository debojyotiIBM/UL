/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package invoice

var PROCESS_INTERNAL_STEP map[string]string
var PROCESS_MAP_STEP map[string]string

func InitiatizeProcessStepLookups() {
	internalStepDescMap := map[string]string{
		"st-GRNSelection-ap1":    "IBM AP - DUPLICATE POLINES DETECTED",
		"st-GRNSelection-ap2":    "IBM AP - POLINES NOT FOUND FOR INVOICE",
		"st-GRNSelection-ap3":    "IBM AP - GOODSRECEIPT FLAG ISSUE",
		"st-GRNSelection-ap4":    "IBM AP - Invoice carrying both 2 and 3 way match POs",
		"st-GRNSelection-ap5":    "IBM AP - GRN REFERENCE BASED GRN SELECTION - INVOICE VALUE AND GRN VALUE MISMATCH",
		"st-GRNSelection-buyer1": "Awaiting GRN – Invoice missing GRN",
		"st-GRNSelection-buyer2": "BUYER - FORWARDED FROM AP",
		"st-GRNSelection-ap6":    "IBM AP - UP NOT MATCHING FOR INVOICE LINES HAVING SAME PO LINES",
		"st-GRNSelection-ap7":    "IBM AP - FORWARDED FROM BUYER",
		"st-GRNSelection-end":    "GRNSelection Stage Ended",
		"st-GRNSelection-1":      "Awaiting GRN",

		"st-VGRN-ap1": "IBM AP - DUPLICATE POLINES DETECTED",
		"st-VGRN-ap2": "IBM AP - POLINES NOT FOUND FOR INVOICE",
		"st-VGRN-ap3": "IBM AP - MATCHAGAINSTGRN FLAG ISSUE",
		"st-VGRN-ap4": "IBM AP - Invoice carrying both VGRN and 3 way match POs",
		"st-VGRN-ap5": "IBM AP - FORWARDED FROM BUYER",
		"st-VGRN-end": "VGRN Stage Ended",

		"ST0000": "New Invoice",
		"ST0001": "Vendor name Match completed",
		"ST0002": "Invoice vendor name or vendor address doesn't match",
		"ST0003": "VMD modification is accepted by the AP team",
		"ST0004": "VMD modification is routed to buyer for approval",
		"ST0005": "AP confirmed vendor modification",
		"ST0006": "VMD modification - rejection is accepted by buyer",
		"ST0007": "VMD modification - buyer returns back to AP",
		"ST0008": "Vendor is blocked",
		"ST0009": "Vendor is deleted",
		"ST0010": "Vendor Master Team confirms Rejection-AP rejects Invoice",
		"ST0011": "Vendor taxId and Invoice taxId Match completed",

		"ST0101": "Vendor ID match completed",
		"ST0102": "Invoice Rejected - Invoice vendor id does not match with PO vendor id",

		"ST0201": "Match Company Code completed",
		"ST0202": "Invoice Rejected company code doesn't match",

		"ST0300": "Invalid PO",
		"ST0301": "Verify PO Status - PO is Deleted / Blocked / Closed",
		"ST0302": "PO Deleted or Blocked or Closed - Buyer or planner Email details missing",
		"ST0303": "Invoice total amount is smaller than or equal to PO Budget",
		"ST0304": "Invoice total amount is greater than PO Budget",
		"ST0305": "AP forwards to buyer for PO fix",
		"ST0306": "PO out of budget - AP Team confirms the PO Fix rejection from buyer and hence rejecting the invoice",
		"ST0307": "Buyer or Planner unblocks the PO",
		"ST0308": "Buyer or Planner provides alternative PO",
		"ST0309": "Buyer chooses to reject the invoice",
		"ST0310": "Vendor modification - Pending with VMD",
		"ST0311": "Buyer or planner returns to AP team for PO status research",

		"ST0401": "Remit to ID found",
		"ST0402": "Vendor not found - payment address missing",
		"ST0403": "Vendor not found - missing bankaccount details",
		"ST0404": "Vendor modification - Buyer confirmed vendor modification",
		"ST0405": "Vendor modification - Buyer returns to AP for research",
		"ST0406": "Vendor modification - Pending with VMD",
		"ST0407": "Vendor modification - Pendign with buyer",
		"ST0408": "Remit to id not found - AP rejects the invoice",
		"ST0409": "Vendor is Blocked from Posting",
		"ST0410": "Vendor is flagged for Deletion",
		"ST0411": "Vendor is flagged for Deletion and Blocked from Posting",
		"ST0412": "Vendor Company Code is Blocked for Posting",
		"ST0413": "Vendor Company Code is Blocked for Payment",
		"ST0414": "Vendor Company Code is Marked for Deletion",
		"ST0415": "Vendor Company Code is Marked for Deletion and Blocked for Posting",
		"ST0416": "Vendor Company Code is Marked for Deletion and Blocked for Payment",
		"ST0417": "Vendor Company Code is Blocked for Posting and Payment",
		"ST0418": "Vendor Company Code is Marked for Deletion, Blocked for Posting and Payment",
		"ST0419": "Vendor is not extended to Purchase Order Company Code",

		"ST0501": "Invoice Passed Post Facto PO Stage",
		"ST0502": "Invoice Rejected - Duplicate Invoice",

		"ST0601": "",
		"ST0602": "Single line PO QTY mismatch - Awaiting for AP Team to select line item",
		"ST0603": "Invoice fix accepted to update line item",
		"ST0604": "Invoice Rejected - AP Team rejects invoice saying Incorrect PO/Line item on invoice",
		"ST0605": "Price Mismatch - Awaiting AP Team for line item selection",
		"ST0606": "",
		"ST0607": "Line item price, quantity and Description matched",
		"ST0608": "Awaiting for AP approval on Line Item Selection description",
		"ST0609": "Invoice Accepted - Buyer/Planner Approved for Line Item",
		"ST0610": "Invoice Rejected - Buyer/Planner rejects the Line Item for invoice",
		"ST0611": "Invoice Re-submit - Buyer/Planner re-submits invoice with alternative PO",
		"ST0612": "Line item price, quantity and Description matched",
		"ST0613": "Multiline PO - Incorrect PO line",

		"ST0701": "No Additional Lines present in invoice",
		"ST0702": "Awaiting Buyer / planner to approve Additional lines",
		"ST0703": "Invoice Accepted - Buyer or Planner Approved additional line items",
		"ST0704": "Invoice Rejected - Buyer or Planner rejected the additional line items",

		"ST0801": "Invoice unit price matches with PO",
		"ST0802": "Awaiting for Buyer/Planner confirmation on unit price",
		"ST0803": "Invoice Accepted - Buyer/Planner Approved Unit price quoted",
		"ST0804": "Invoice Rejected - Buyer/Planner rejected the unit price",

		"ST0901": "GRN exists for the invoice",
		"ST0902": "Waiting for GRN to be created",
		"ST0903": "Buyer needs to approve the GRN quantity mismatch",
		"ST0904": "Buyer needs to approve one or more GRN quantity mismatch",
		"ST0905": "Invoice line pushed to easy robo queue",
		"ST0906": "Invoice rejected by Buyer",
		"ST0907": "BOL mismatch needs to be approved by buyer",
		"ST0908": "Alternate GRN provided by buyer",
		"ST0909": "Invoice is on hold as one or more invoice line is still waiting for GRN",

		"ST01001": "Tax validated. Invoice processed completely",
		"ST01002": "Invoice Rejected - Incorrect tax",

		"ST9999":  "Invoice Rejected In BCI Preprocessing",
		"ST01102": "Reportabilty Match completed",
		"ST01103": "Reportabilty routed to buyer for approval",
		"ST01104": "Reportabilty approved by buyer",
		"ST01105": "Reportabilty reject by buyer and ready for manual posting",
		"ST01201": "No Duplicate found for the invoice",
		"ST01202": "Invoice Rejected in Post Facto PO Stage",
	}

	processMapStep := map[string]string{
		"ST0000": "",
		"ST0001": "1A",
		"ST0002": "1A",
		"ST0003": "1B",
		"ST0004": "1G",
		"ST0005": "1C",
		"ST0006": "1H",
		"ST0007": "1H",
		"ST0010": "1H",
		"ST0011": "17",

		"ST0101": "4",
		"ST0102": "1H",

		"ST0201": "5a",
		"ST0202": "1H",

		"ST0301": "6",
		"ST0302": "6a.7",
		"ST0303": "7a",
		"ST0304": "7a",
		"ST0305": "9a",
		"ST0306": "9a.1",
		"ST0307": "8a.2",
		"ST0308": "8a.3",
		"ST0309": "8a.2",
		"ST0310": "8a.8",
		"ST0311": "8a.4",

		"ST0401": "10.c",
		"ST0402": "10.c",
		"ST0403": "10.c",
		"ST0404": "10",
		"ST0405": "10",
		"ST0406": "10c.1",
		"ST0407": "10E",
		"ST0408": "10c.1a1",

		"ST0501": "11b",
		"ST0502": "11b1",

		"ST0601": "",
		"ST0602": "12.e1a",
		"ST0603": "12e1.a",
		"ST0604": "12H",
		"ST0605": "12.f.1a",
		"ST0606": "",
		"ST0607": "12.i.1",
		"ST0608": "12j",
		"ST0609": "12j.1",
		"ST0610": "12j.1b",
		"ST0611": "12j.1c",
		"ST0612": "12.i.1",
		"ST0613": "12.d",

		"ST0701": "13a",
		"ST0702": "13c",
		"ST0703": "13c1",
		"ST0704": "13d",

		"ST0801": "14.a.1",
		"ST0802": "14c",
		"ST0803": "14c.1",
		"ST0804": "14d",

		"ST0901": "15a",
		"ST0902": "15a",
		"ST0903": "15",
		"ST0904": "15",
		"ST0905": "15",
		"ST0906": "15",
		"ST0907": "15",
		"ST0908": "15",
		"ST0909": "15",

		"ST01001": "16",
		"ST01002": "16",

		"ST9999":  "0.4.1",
		"ST01102": "16",
		"ST01103": "16a",
		"ST01104": "16b",
		"ST01105": "16c",
		"ST01201": "17",
		"ST01202": "17a",
	}

	PROCESS_INTERNAL_STEP = internalStepDescMap
	PROCESS_MAP_STEP = processMapStep
}
