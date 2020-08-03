
package com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.dto.docextraction;

import com.fasterxml.jackson.annotation.*;
import lombok.Data;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@JsonInclude(JsonInclude.Include.NON_NULL)
@Data
public class DocumentExtractResult {

    @JsonProperty("status")
    private String status;
    @JsonProperty("claimRequestId")
    private Long claimRequestId;
    @JsonProperty("documentType")
    private String documentType;
    @JsonProperty("languages")
    private List<String> languages = null;
    @JsonProperty("country")
    private String country;
    @JsonProperty("created")
    private String created;
    @JsonProperty("extraction")
    private Extraction extraction;
    @JsonIgnore
    private Map<String, Object> additionalProperties = new HashMap<String, Object>();
}
