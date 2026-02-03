import {
  ComponentsComponent,
  ComponentsNode,
  CanvasesCanvasNodeExecution,
  CanvasesCanvasNodeQueueItem,
} from "@/api-client";
import { ComponentBaseMapper, OutputPayload } from "../../types";
import { ComponentBaseProps, EventSection } from "@/ui/componentBase";
import { getBackgroundColorClass, getColorClass } from "@/utils/colors";
import { getState, getStateMap, getTriggerRenderer } from "../..";
import awsIcon from "@/assets/icons/integrations/aws.svg";
import { formatTimeAgo } from "@/utils/date";
import { MetadataItem } from "@/ui/metadataList";
import { CodeArtifactPackageVersionDescription, CodeArtifactTriggerConfiguration, CodeArtifactTriggerMetadata } from "./types";
import { buildCodeArtifactMetadataItems, formatPackageName } from "./utils";
import { numberOrZero, stringOrDash } from "../../utils";

export const describePackageVersionMapper: ComponentBaseMapper = {
  props(
    nodes: ComponentsNode[],
    node: ComponentsNode,
    componentDefinition: ComponentsComponent,
    lastExecutions: CanvasesCanvasNodeExecution[],
    _items?: CanvasesCanvasNodeQueueItem[],
  ): ComponentBaseProps {
    const lastExecution = lastExecutions.length > 0 ? lastExecutions[0] : null;
    const componentName = componentDefinition.name || node.component?.name || "unknown";

    return {
      title: node.name || componentDefinition.label || componentDefinition.name || "Unnamed component",
      iconSrc: awsIcon,
      iconColor: getColorClass(componentDefinition.color),
      collapsedBackground: getBackgroundColorClass(componentDefinition.color),
      collapsed: node.isCollapsed,
      eventSections: lastExecution ? getDescribePackageVersionEventSections(nodes, lastExecution, componentName) : undefined,
      includeEmptyState: !lastExecution,
      metadata: getDescribePackageVersionMetadataList(node),
      eventStateMap: getStateMap(componentName),
    };
  },

  getExecutionDetails(execution: CanvasesCanvasNodeExecution, _node: ComponentsNode): Record<string, string> {
    const outputs = execution.outputs as { default?: OutputPayload[] } | undefined;
    const result = outputs?.default?.[0]?.data as CodeArtifactPackageVersionDescription | undefined;

    if (!result) {
      return {};
    }

    const licenses = result.licenses?.map((license) => license.name).filter(Boolean) || [];

    return {
      Package: stringOrDash(formatPackageName(result.namespace ?? undefined, result.packageName)),
      Version: stringOrDash(result.version),
      Status: stringOrDash(result.status),
      Format: stringOrDash(result.format),
      Revision: stringOrDash(result.revision),
      "Display Name": stringOrDash(result.displayName),
      "Published Time": result.publishedTime ? new Date(result.publishedTime * 1000).toISOString() : "-",
      "Origin Type": stringOrDash(result.origin?.originType),
      "Origin Repository": stringOrDash(result.origin?.domainEntryPoint?.repositoryName),
      "Origin Connection": stringOrDash(result.origin?.domainEntryPoint?.externalConnectionName),
      Licenses: licenses.length > 0 ? licenses.join(", ") : "-",
      "License Count": numberOrZero(result.licenses?.length).toString(),
      Summary: stringOrDash(result.summary),
      "Source Code": stringOrDash(result.sourceCodeRepository),
      "Home Page": stringOrDash(result.homePage),
    };
  },

  subtitle(_node: ComponentsNode, execution: CanvasesCanvasNodeExecution): string {
    if (!execution.createdAt) {
      return "";
    }
    return formatTimeAgo(new Date(execution.createdAt));
  },
};

function getDescribePackageVersionMetadataList(node: ComponentsNode): MetadataItem[] {
  const nodeMetadata = node.metadata as CodeArtifactTriggerMetadata | undefined;
  const configuration = node.configuration as CodeArtifactTriggerConfiguration | undefined;
  return buildCodeArtifactMetadataItems(nodeMetadata, configuration);
}

function getDescribePackageVersionEventSections(
  nodes: ComponentsNode[],
  execution: CanvasesCanvasNodeExecution,
  componentName: string,
): EventSection[] {
  const rootTriggerNode = nodes.find((n) => n.id === execution.rootEvent?.nodeId);
  const rootTriggerRenderer = getTriggerRenderer(rootTriggerNode?.trigger?.name || "");
  const { title } = rootTriggerRenderer.getTitleAndSubtitle(execution.rootEvent!);

  return [
    {
      receivedAt: new Date(execution.createdAt!),
      eventTitle: title,
      eventSubtitle: formatTimeAgo(new Date(execution.createdAt!)),
      eventState: getState(componentName)(execution),
      eventId: execution.rootEvent!.id!,
    },
  ];
}
