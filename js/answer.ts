// @ts-nocheck
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
class AnswerSheetDetect {
    getCenterSortY(innerPoints) {
        // y坐标排序
        let contour_with_centers = [];
        for (let i = 0; i < innerPoints.length; i++) {
            // 计算轮廓的矩
            let moments = cv.moments(innerPoints[i]);
            // 计算中心坐标
            let cx = moments.m10 / moments.m00;
            let cy = moments.m01 / moments.m00;
            // 将轮廓和中心坐标存储在 innerPoints 数组中
            contour_with_centers.push({ x: cx, y: cy });
        }
        contour_with_centers.sort((a, b) => a.y - b.y);
        return contour_with_centers;
    }
    countBlack(image) {
        let blackPixelCount = 0;
        for (let row = 0; row < image.rows; row++) {
            for (let col = 0; col < image.cols; col++) {
                let pixel = image.ucharAt(row, col); // 获取灰度值
                if (pixel === 0) { // 黑色像素
                    blackPixelCount++;
                }
            }
        }
        return blackPixelCount;
    }
    ellipsePerimeter(a, b) {
        // Calculate the eccentricity squared
        let h = Math.pow((a - b), 2) / Math.pow((a + b), 2);
        // Calculate the perimeter approximation
        return Math.PI * (a + b) * (1 + (3 * h) / (10 + Math.sqrt(4 - 3 * h)));
    }
    resize(input, width, height) {
        let output = new cv.Mat();
        cv.resize(input, output, new cv.Size(width, height));
        return output;
    }
    fourPointTransform(image_mat, points) {
        let br = points.reduce((prev, current) => (prev.x + prev.y > current.x + current.y) ? prev : current);
        let tl = points.reduce((prev, current) => (prev.x + prev.y < current.x + current.y) ? prev : current);
        let other = points.filter(obj => (obj.x !== br.x || obj.y !== br.y) && (obj.x !== tl.x || obj.y !== tl.y));
        let bl = other.reduce((prev, current) => (prev.y > current.y) ? prev : current);
        let tr = other.reduce((prev, current) => (prev.x > current.x) ? prev : current);
        let widthA = Math.sqrt(Math.pow((br.x - bl.x), 2.0) + Math.pow((br.y - bl.y), 2.0));
        let widthB = Math.sqrt(Math.pow((tr.x - tl.x), 2.0) + Math.pow((tr.y - tl.y), 2.0));
        let heightA = Math.sqrt(Math.pow((tr.x - br.x), 2.0) + Math.pow((tr.y - br.y), 2.0));
        let heightB = Math.sqrt(Math.pow((tl.x - bl.x), 2.0) + Math.pow((tl.y - bl.y), 2.0));
        let maxHeight = heightA > heightB ? heightA : heightB;
        let maxWidth = widthA > widthB ? widthA : widthB;
        let srcTri = cv.matFromArray(4, 1, cv.CV_32FC2, [tl.x, tl.y, tr.x, tr.y, br.x, br.y, bl.x, bl.y]);
        let dstTri = cv.matFromArray(4, 1, cv.CV_32FC2, [0, 0, maxWidth - 1, 0, maxWidth - 1, maxHeight - 1, 0, maxHeight]);
        let M = cv.getPerspectiveTransform(srcTri, dstTri);
        let warped = new cv.Mat();
        let dsize = new cv.Size(maxWidth, maxHeight);
        cv.warpPerspective(image_mat, warped, M, dsize, cv.INTER_LINEAR, cv.BORDER_CONSTANT);
        M.delete();
        return warped;
        // return this.resize(warped,248 * 4, 351 * 4)
    }
    groupAndSort(centers) {
        const sortedGroups = [];
        // Iterate over the contours in groups of 5
        for (let i = 0; i < centers.length; i += 5) {
            // Get the current group of up to 5 elements
            const group = centers.slice(i, i + 5);
            // Sort the group by x-coordinate
            group.sort((a, b) => a.x - b.x);
            // Add the sorted group to the sortedGroups array
            sortedGroups.push(...group);
        }
        return sortedGroups;
    }
    isSimilar(coord1, coord2, threshold) {
        return Math.abs(coord1.x - coord2.x) < threshold && Math.abs(coord1.y - coord2.y) < threshold;
    }
    getOptionsAns(j) {
        switch (j) {
            case 0:
                return "A";
            case 1:
                return "B";
            case 2:
                return "C";
            case 3:
                return "D";
        }
    }
    isContourCircle(contour) {
        let area = cv.contourArea(contour);
        let perimeter = cv.arcLength(contour, true);
        if (perimeter === 0) {
            return false;
        }
        let circularity = 4 * Math.PI * (area / (perimeter * perimeter));
        return 0.8 < circularity && circularity < 1.2;
    }
    removeSimilarCoordinates(all_center, threshold) {
        let uniqueCenters = [];
        for (let center of all_center) {
            let isUnique = true;
            for (let uniqueCenter of uniqueCenters) {
                if (this.isSimilar(center, uniqueCenter, threshold)) {
                    isUnique = false;
                    break;
                }
            }
            if (isUnique) {
                uniqueCenters.push(center);
            }
        }
        return uniqueCenters;
    }
    findSimilarX(uniqueCenters, findX) {
        let group = [];
        for (let item of uniqueCenters) {
            if (Math.abs(item.x - findX) < 8) {
                group.push(item);
            }
        }
        return group;
    }
    findSimilarY(uniqueCenters, findY) {
        let group = [];
        for (let item of uniqueCenters) {
            if (Math.abs(item.y - findY) < 10) {
                group.push(item);
            }
        }
        return group;
    }
    validatePhoneNumber(phoneNumber) {
        const phoneRegex = /^1[3-9]\d{9}$/;
        return phoneRegex.test(phoneNumber);
    }
    getAvgW(uniqueCenters, diffX, diffY) {
        uniqueCenters.sort((a, b) => a.y - b.y);
        let newGroup = [];
        let curr = 0;
        let i = 1;
        while (i < uniqueCenters.length) {
            let tmp = [];
            tmp.push(uniqueCenters[curr]);
            // while (i < uniqueCenters.length && Math.abs(uniqueCenters[curr].y - uniqueCenters[i].y) < 10){
            while (i < uniqueCenters.length && Math.abs(uniqueCenters[curr].y - uniqueCenters[i].y) < diffY) {
                tmp.push(uniqueCenters[i]);
                i++;
            }
            newGroup.push(tmp);
            curr = i;
            i++;
        }
        let allW = 0;
        let wCount = 0;
        for (let i = 0; i < newGroup.length; i++) {
            let tmpGroup = newGroup[i];
            tmpGroup.sort((a, b) => a.x - b.x);
            if (tmpGroup.length < 2) {
                continue;
            }
            let pre = 0;
            let curr = 1;
            while (curr < tmpGroup.length) {
                // if (Math.abs(tmpGroup[pre].x - tmpGroup[curr].x) < 30){
                if (Math.abs(tmpGroup[pre].x - tmpGroup[curr].x) < diffX) {
                    allW += Math.abs(tmpGroup[pre].x - tmpGroup[curr].x);
                    wCount++;
                }
                pre++;
                curr++;
            }
        }
        return Math.round(allW / wCount * 10) / 10;
    }
    getAvgH(uniqueCenters, diffX, diffY) {
        uniqueCenters.sort((a, b) => a.x - b.x);
        let newGroup = [];
        let curr = 0;
        let i = 1;
        while (i < uniqueCenters.length) {
            let tmp = [];
            tmp.push(uniqueCenters[curr]);
            // while (i < uniqueCenters.length && Math.abs(uniqueCenters[curr].x - uniqueCenters[i].x) < 5){
            while (i < uniqueCenters.length && Math.abs(uniqueCenters[curr].x - uniqueCenters[i].x) < diffX) {
                tmp.push(uniqueCenters[i]);
                i++;
            }
            newGroup.push(tmp);
            curr = i;
            i++;
        }
        let allH = 0;
        let hCount = 0;
        for (let i = 0; i < newGroup.length; i++) {
            let tmpGroup = newGroup[i];
            tmpGroup.sort((a, b) => a.y - b.y);
            if (tmpGroup.length < 2) {
                continue;
            }
            let pre = 0;
            let curr = 1;
            while (curr < tmpGroup.length) {
                // if (Math.abs(tmpGroup[pre].y - tmpGroup[curr].y) < 23){
                if (Math.abs(tmpGroup[pre].y - tmpGroup[curr].y) < diffY) {
                    allH += Math.abs(tmpGroup[pre].y - tmpGroup[curr].y);
                    hCount++;
                }
                pre++;
                curr++;
            }
        }
        return Math.round(allH / hCount * 10) / 10;
    }
    completePhoneArea(uniqueCenters, dstW, dstH) {
        uniqueCenters.sort((a, b) => a.y - b.y);
        let defaultY = uniqueCenters[0].y;
        uniqueCenters.sort((a, b) => a.x - b.x);
        let defaultX = uniqueCenters[0].x;
        let avgW = this.getAvgW(uniqueCenters, dstW / 11, 10);
        let avgH = this.getAvgH(uniqueCenters, 10, dstH / 11);
        // console.log(avgH,avgW,defaultY,defaultX)
        let res = [];
        for (let i = 0; i < 11; i++) {
            let findX = defaultX + avgW * i;
            let groupX = this.findSimilarX(uniqueCenters, findX);
            if (groupX.length == 0) {
                groupX.push({ x: findX, y: defaultY });
            }
            groupX.sort((a, b) => a.y - b.y);
            if (groupX.length == 10) {
                res.push(groupX);
                continue;
            }
            let pre = groupX[0];
            let step = 0;
            // 补全 pre 前面的
            if (!(this.isSimilar(pre, { x: findX, y: defaultY }, 5))) {
                // 计算按 y 坐标排序后的第一位,是不是第一个元素,不是的话按 num 补全curr前面的元素
                step = Math.round(Math.abs(pre.y - defaultY) / avgH);
                for (let j = 1; j <= step; j++) {
                    let appendTmp = { x: pre.x, y: pre.y - j * avgH };
                    groupX.unshift(appendTmp);
                }
            }
            if (groupX.length == 10) {
                res.push(groupX);
                continue;
            }
            let curr = step; // 这里 curr 一定不是第 10 位,所以还需要补
            // 如果 curr 是 group_x 最后一个元素,一直补齐到第 10 个就行了
            while (curr < 9) {
                if (curr == groupX.length - 1) {
                    let currItem = groupX[curr];
                    for (let j = 1; j < 10 - curr; j++) {
                        let appendTmp = { x: currItem.x, y: currItem.y + j * avgH };
                        groupX.push(appendTmp);
                    }
                    break;
                }
                else {
                    // 这里 next 一定存在于 group_x 中,因为上个判断已经判断了
                    let next = curr + 1; // 如果 next 存在于 group_x 中,就补全 curr 和 next 之间的元素,如果不需要补,就 continue,直到 curr == len(group_x)-1
                    let next_item = groupX[next];
                    let curr_item = groupX[curr];
                    let num = Math.round(Math.abs(next_item.y - curr_item.y) / avgH);
                    if (num > 1) {
                        let preGroup = groupX.slice(0, curr + 1);
                        let lastGroup = groupX.slice(next, groupX.length);
                        let tmpGroup = [];
                        for (let j = 1; j < num; j++) {
                            let tmpAppend = { x: curr_item.x, y: curr_item.y + j * avgH };
                            tmpGroup.push(tmpAppend);
                        }
                        groupX = [...preGroup, ...tmpGroup, ...lastGroup];
                    }
                    curr = curr + num;
                }
            }
            res.push(groupX);
        }
        return res;
    }
    getPhone(sortedGroups, warped) {
        let phonePoints = [];
        phonePoints.push(sortedGroups[2]);
        phonePoints.push(sortedGroups[4]);
        phonePoints.push(sortedGroups[7]);
        phonePoints.push(sortedGroups[9]);
        let phoneArea = this.fourPointTransform(warped, phonePoints);
        let last = Math.floor(phoneArea.rows / 4);
        // 定义感兴趣的区域（ROI）
        let rect = new cv.Rect(0, last, phoneArea.cols, phoneArea.rows - last);
        let dst = this.resize(phoneArea.roi(rect), warped.size().width * 0.32, warped.size().height * 0.22);
        let phone = new cv.Mat();
        // let phoneSelect = new cv.Mat()
        let color = new cv.Mat();
        cv.cvtColor(dst, color, cv.COLOR_RGBA2GRAY);
        // cv.Canny(color,canny,10,255)
        cv.adaptiveThreshold(color, phone, 255, cv.ADAPTIVE_THRESH_GAUSSIAN_C, cv.THRESH_BINARY, 33, 1);
        // cv.adaptiveThreshold(color, phoneSelect, 255, cv.ADAPTIVE_THRESH_GAUSSIAN_C, cv.THRESH_BINARY, 33, 1)
        let contours = new cv.MatVector();
        let hierarchy = new cv.Mat();
        // Find contours
        cv.findContours(phone, contours, hierarchy, cv.RETR_TREE, cv.CHAIN_APPROX_SIMPLE);
        let all_center = [];
        let center2contour = new Map();
        let allWidth = 0;
        let allHeight = 0;
        for (let i = 0; i < contours.size(); i++) {
            let contour = contours.get(i);
            let minAreaRect = cv.minAreaRect(contour);
            let maxW = dst.size().width / 10;
            let maxH = dst.size().height / 10;
            if (minAreaRect.size.width > maxW || minAreaRect.size.height > maxH) {
                continue;
            }
            // cv.drawContours(dst, contours, i, [255, 0, 0,255], 1);
            let area = cv.contourArea(contour);
            if (area > maxW * maxH * 0.8 || area < maxW * maxH * 0.25) {
                continue;
            }
            all_center.push(minAreaRect.center);
            center2contour.set(minAreaRect.center, contour);
            allWidth += minAreaRect.size.width;
            allHeight += minAreaRect.size.height;
            // cv.drawContours(dst, contours, i, [0, 255, 0, 255], 1);
        }
        color.delete()
        hierarchy.delete()
        contours.delete()

        if (all_center.length == 0) {
            phone.delete()
            return "";
        }
        // unique
        let uniqueCenter = this.removeSimilarCoordinates(all_center, 5);
        // console.log("uniqueCenter",uniqueCenter)
        // 补全
        let completeCenter = this.completePhoneArea(uniqueCenter, dst.size().width, dst.size().height);
        // console.log("completeCenter",completeCenter)
        if (completeCenter == false) {
            phone.delete()
            return "";
        }
        let defaultWeight = Math.round(allWidth / center2contour.size * 10) / 10;
        let defaultHeight = Math.round(allHeight / center2contour.size * 10) / 10;
        let phoneStr = "";
        for (let i = 0; i < completeCenter.length; i++) {
            let tmpAns = [];
            for (let j = 0; j < completeCenter[i].length; j++) {
                let contour = new cv.Mat();
                // if (completeCenter[i][j] in center2contour){
                if (center2contour.has(completeCenter[i][j])) {
                    contour = center2contour.get(completeCenter[i][j]);
                }
                else {
                    let topLeft = new cv.Point(completeCenter[i][j].x - defaultWeight / 2, completeCenter[i][j].y - defaultHeight / 2);
                    let topRight = new cv.Point(completeCenter[i][j].x + defaultWeight / 2, completeCenter[i][j].y - defaultHeight / 2);
                    let bottomRight = new cv.Point(completeCenter[i][j].x + defaultWeight / 2, completeCenter[i][j].y + defaultHeight / 2);
                    let bottomLeft = new cv.Point(completeCenter[i][j].x - defaultWeight / 2, completeCenter[i][j].y + defaultHeight / 2);
                    // 创建一个 cv.Mat 来存储这些顶点
                    contour = cv.matFromArray(4, 1, cv.CV_32SC2, [
                        topLeft.x, topLeft.y,
                        topRight.x, topRight.y,
                        bottomRight.x, bottomRight.y,
                        bottomLeft.x, bottomLeft.y
                    ]);
                    let rect = cv.boundingRect(contour);
                    let sp1 = new cv.Point(rect.x, rect.y);
                    let sp2 = new cv.Point(rect.x + rect.width, rect.y + rect.height);
                    cv.rectangle(dst, sp1, sp2, [0, 0, 255, 255], 1);
                }
                let rect = cv.boundingRect(contour);
                // 裁剪图像，仅保留轮廓区域
                let cropped_result = phone.roi(rect);
                let blackPixelCount = this.countBlack(cropped_result);
                // 计算黑色像素数
                // console.log(blackPixelCount)
                tmpAns.push(blackPixelCount);
                contour.delete()
            }
            let max = tmpAns[0];
            let index = 0;
            for (let z = 0; z < tmpAns.length; z++) {
                if (tmpAns[z] > max) {
                    max = tmpAns[z];
                    index = z;
                }
            }
            // console.log("---------")
            tmpAns = [];
            phoneStr = phoneStr + index.toString();
        }
        phone.delete()
        // console.log(phoneStr)
        // return [dst,phoneSelect]
        return phoneStr;
    }
    getAnswer(sortedGroups, warped) {
        let num = 1;
        let answers = new Map();
        // 最后 5 个不处理 因为是最后的标记
        for (let i = 0; i < sortedGroups.length - 5; i++) {
            if (i + 6 > sortedGroups.length - 1) {
                break;
            }
            if (i == 4 || i == 9 || i == 14 || i == 19 || i == 24 || i == 29) {
                continue;
            }
            let points = [];
            points.push(sortedGroups[i]);
            points.push(sortedGroups[i + 1]);
            points.push(sortedGroups[i + 5]);
            points.push(sortedGroups[i + 6]);
            let ansArea = this.fourPointTransform(warped, points);
            let resize = this.resize(ansArea, warped.size().width / 5, warped.size().height / 8);
            let last = Math.floor(resize.rows / 5);
            // 定义感兴趣的区域（ROI）
            let rect = new cv.Rect(0, last, resize.cols, resize.rows - last);
            let dst = resize.roi(rect);
            let ans = [];
            try {
                ans = this.dealAnswer(dst)
            }catch (e) {
                console.log("dealAnswer err:"+e);
            }
            // 5 道题
            for (let j = 0; j < 5; j++) {
                if (j < ans.length) {
                    answers.set(num, ans[j]);
                }
                else {
                    answers.set(num, []);
                }
                num++;
            }
        }
        return answers;
    }
    completeAnswer(uniqueCenters, dstW, dstH) {
        let default_y = dstW * 0.078;
        let default_x = dstH * 0.29;
        uniqueCenters.sort((a, b) => a.y - b.y);
        if (uniqueCenters[0].y < 45) {
            let group = this.findSimilarY(uniqueCenters, uniqueCenters[0].y);
            let all = 0;
            for (let i = 0; i < group.length; i++) {
                all += group[i].y;
            }
            default_y = Math.round(all / group.length * 10) / 10;
        }
        uniqueCenters.sort((a, b) => a.x - b.x);
        if (uniqueCenters[0].x < 70) {
            let group = this.findSimilarX(uniqueCenters, uniqueCenters[0].x);
            let all = 0;
            for (let i = 0; i < group.length; i++) {
                all += group[i].x;
            }
            default_x = Math.round(all / group.length * 10) / 10;
        }
        // console.log(default_x,default_y)
        // 圆心直接的平均宽度距离
        let avg_w = this.getAvgW(uniqueCenters, dstW / 5, 10);
        // 圆心直接的平均高度距离
        let avg_h = this.getAvgH(uniqueCenters, 10, dstH / 4.5);
        // console.log("aaaa",default_x,default_y,avg_w,avg_h)
        uniqueCenters.sort((a, b) => a.y - b.y);
        let res = [];
        for (let i = 0; i < 5; i++) {
            let findX = default_x + avg_w * i;
            let groupX = this.findSimilarX(uniqueCenters, findX);
            if (groupX.length == 0) {
                groupX.push({ x: findX, y: default_y });
            }
            if (groupX.length == 4) {
                res.push(groupX);
                continue;
            }
            groupX.sort((a, b) => a.y - b.y);
            let pre = groupX[0];
            let step = 0;
            if (!(this.isSimilar(pre, { x: findX, y: default_y }, 5))) {
                // 计算按 y 坐标排序后的第一位,是不是第一个元素,不是的话按 num 补全curr前面的元素
                step = Math.round(Math.abs(pre.y - default_y) / avg_h);
                for (let z = 1; z <= step; z++) {
                    groupX.unshift({ x: pre.x, y: pre.y - z * avg_h });
                }
            }
            if (groupX.length == 4) {
                res.push(groupX);
                continue;
            }
            let curr = step;
            while (curr < 3) {
                if (curr == groupX.length - 1) {
                    let currItem = groupX[curr];
                    for (let z = 1; z < 4 - curr; z++) {
                        groupX.push({ x: currItem.x, y: currItem.y + z * avg_h });
                    }
                    break;
                }
                else {
                    let next = curr + 1;
                    let nextItem = groupX[next];
                    let currItem = groupX[curr];
                    let num = Math.round(Math.abs(nextItem.y - currItem.y) / avg_h);
                    if (num > 1) {
                        let preGroup = groupX.slice(0, curr + 1);
                        let lastGroup = groupX.slice(next, groupX.length);
                        let tmpGroup = [];
                        for (let j = 1; j < num; j++) {
                            tmpGroup.push({ x: currItem.x, y: currItem.y + j * avg_h });
                        }
                        groupX = [...preGroup, ...tmpGroup, ...lastGroup];
                    }
                    curr = curr + num;
                }
            }
            res.push(groupX);
        }
        return res;
    }
    dealAnswer(dst) {
        let ans = new cv.Mat();
        let ansSelect = new cv.Mat();
        let color = new cv.Mat();
        let resAnswer = [];
        try {
            cv.cvtColor(dst, color, cv.COLOR_RGBA2GRAY);
            cv.adaptiveThreshold(color, ans, 255, cv.ADAPTIVE_THRESH_MEAN_C, cv.THRESH_BINARY, 33, 1);
            cv.adaptiveThreshold(color, ansSelect, 255, cv.ADAPTIVE_THRESH_MEAN_C, cv.THRESH_BINARY, 111, 11);
        }catch (e){
            console.log("adaptiveThreshold err:" + e);
            color.delete()
            ansSelect.delete()
            ans.delete()
            return resAnswer
        }

        let contours = new cv.MatVector();
        let hierarchy = new cv.Mat();
        // Find contours
        cv.findContours(ans, contours, hierarchy, cv.RETR_TREE, cv.CHAIN_APPROX_SIMPLE);
        let all_center = [];
        let center2contour = new Map();
        let allW = 0;
        let allH = 0;
        for (let i = 0; i < contours.size(); i++) {
            let contour = contours.get(i);
            let minAreaRect = cv.minAreaRect(contour);
            // console.log("ddd",minAreaRect.size.width,minAreaRect.size.height)
            if (minAreaRect.size.width > dst.size().width / 6 || minAreaRect.size.height > dst.size().height / 4.5) {
                continue;
            }
            if (minAreaRect.size.width < dst.size().width / 8 || minAreaRect.size.height < dst.size().height / 6.6) {
                continue;
            }
            allW += minAreaRect.size.width;
            allH += minAreaRect.size.height;
            all_center.push(minAreaRect.center);
            center2contour.set(minAreaRect.center, contour);
            // cv.drawContours(dst, contours, i, [255, 0, 0, 255], 1);
        }
        color.delete()
        contours.delete()
        hierarchy.delete()
        ansSelect.delete()
        if (all_center.length == 0) {
            ansSelect.delete()
            return [];
        }
        // console.log(all_center)
        // 轮廓去重
        let uniqueCenters = this.removeSimilarCoordinates(all_center, 5);
        // console.log(uniqueCenters)
        let completeCenters = this.completeAnswer(uniqueCenters, dst.size().width, dst.size().height);
        // console.log(completeCenters)
        let defaultWeight = Math.round(allW / center2contour.size * 10) / 10;
        let defaultHeight = Math.round(allH / center2contour.size * 10) / 10;
        // console.log(defaultWeight, defaultHeight)
        // resAnswer.push(ansSelect)
        for (let groupIndex = 0; groupIndex < completeCenters.length; groupIndex++) {
            let group = completeCenters[groupIndex];
            let tmpBack = [];
            let groupAnsList = [];
            for (let j = 0; j < group.length; j++) {
                let contour = null;
                if (center2contour.has(group[j])) {
                    contour = center2contour.get(group[j]);
                } else {
                    let topLeft = new cv.Point(group[j].x - defaultWeight / 2, group[j].y - defaultHeight / 2);
                    let topRight = new cv.Point(group[j].x + defaultWeight / 2, group[j].y - defaultHeight / 2);
                    let bottomRight = new cv.Point(group[j].x + defaultWeight / 2, group[j].y + defaultHeight / 2);
                    let bottomLeft = new cv.Point(group[j].x - defaultWeight / 2, group[j].y + defaultHeight / 2);
                    // 创建一个 cv.Mat 来存储这些顶点
                    contour = cv.matFromArray(4, 1, cv.CV_32SC2, [
                        topLeft.x, topLeft.y,
                        topRight.x, topRight.y,
                        bottomRight.x, bottomRight.y,
                        bottomLeft.x, bottomLeft.y
                    ]);
                    // let rect = cv.boundingRect(contour)
                    // let sp1 = new cv.Point(rect.x, rect.y)
                    // let sp2 = new cv.Point(rect.x + rect.width, rect.y + rect.height)
                    // cv.rectangle(dst, sp1, sp2, [0, 0, 255, 255], 2)
                }
                try {
                    // let l = new cv.Mat()
                    // cv.approxPolyDP(contour,l,0.025 * cv.arcLength(contour, true),true)
                    let rect = cv.boundingRect(contour);
                    // console.log(rect)
                    let w = Math.round(rect.width / 8 * 10) / 10;
                    let h = Math.round(rect.height / 7 * 10) / 10;
                    rect.x = rect.x + w;
                    rect.y = rect.y + h;
                    rect.width = rect.width - w * 1.8;
                    rect.height = rect.height - h * 1.8;
                    // console.log("after",rect.x, rect.y, rect.width, rect.height)
                    let optionsImage = ans.roi(rect);
                    let resizeImage = this.resize(optionsImage, defaultWeight, defaultHeight);
                    // imageArr.push(resizeImage)
                    let blackPixelCount = this.countBlack(resizeImage);
                    // let total = resizeImage.width * resizeImage.height
                    let total = defaultWeight * defaultHeight;
                    if (Math.round(blackPixelCount / total * 10000) / 10000 > 0.55) {
                        // console.log('blackPixelCount', Math.round(blackPixelCount/total * 10000) / 10000)
                        tmpBack.push(this.getOptionsAns(j));
                    }
                }catch (e) {
                    console.log("deal group answer failed: ", e.toString());
                }
            }
            resAnswer.push(tmpBack);
            // console.log(tmpBack);
        }
        ans.delete()
        // return dst
        // return [resAnswer,imageArr]
        return resAnswer;
    }
    getInnerPoint(image_mat, points) {
        // Detect contours
        let contours = new cv.MatVector();
        let hierarchy = new cv.Mat();
        let warpedGray = new cv.Mat();
        let warped
        try {
            warped = this.fourPointTransform(image_mat, points);
            // Convert to grayscale and apply adaptive thresholding
            cv.cvtColor(warped, warpedGray, cv.COLOR_RGBA2GRAY, 0);
            cv.GaussianBlur(warpedGray, warpedGray, new cv.Size(5, 5), 0, 0);
            // cv.Canny(warpedGray, warpedGray, 10, 400)
            cv.adaptiveThreshold(warpedGray, warpedGray, 255, cv.ADAPTIVE_THRESH_GAUSSIAN_C, cv.THRESH_BINARY, 11, 2);
            cv.findContours(warpedGray, contours, hierarchy, cv.RETR_TREE, cv.CHAIN_APPROX_SIMPLE);
        }catch (e) {
            console.log("getInnerPoint findContours failed: ", e.toString());
            return [[],[]]
        }
        let innerPoints = [];
        for (let i = 0; i < contours.size(); i++) {
            let k = i;
            let ic = 0;
            while (hierarchy.intPtr(0, k)[2] !== -1) {
                k = hierarchy.intPtr(0, k)[2];
                ic++;
            }
            if (ic === 2 && this.isContourCircle(contours.get(i))) {
                let ellipse = cv.fitEllipse(contours.get(i));
                let area = Math.PI * ellipse.size.width * ellipse.size.height / 4;
                if (area > 2000) {
                    continue;
                }
                let arcEll = this.ellipsePerimeter(ellipse.size.width, ellipse.size.height);
                let arcOri = cv.arcLength(contours.get(i), true) * 2;
                if (0.90 <= cv.contourArea(contours.get(i)) / area && cv.contourArea(contours.get(i)) / area <= 1.1 &&
                    0.90 <= arcEll / arcOri && arcEll / arcOri <= 1.1) {
                    innerPoints.push(contours.get(i));
                    // cv.drawContours(warped, contours, i, [0, 0, 255,255], 2);
                }
            }
        }
        hierarchy.delete()
        warpedGray.delete()
        contours.delete()
        return [innerPoints, warped];
    }
    getScreenContour(mat) {
        // Convert to grayscale
        let gray = new cv.Mat();
        cv.cvtColor(mat, gray, cv.COLOR_BGR2GRAY);
        cv.GaussianBlur(gray, gray, new cv.Size(3, 3), 0, 0);
        cv.adaptiveThreshold(gray, gray, 255, cv.ADAPTIVE_THRESH_MEAN_C, cv.THRESH_BINARY, 11, 2);
        let contours = new cv.MatVector();
        let hierarchy = new cv.Mat();
        // Find contours
        cv.findContours(gray, contours, hierarchy, cv.RETR_TREE, cv.CHAIN_APPROX_SIMPLE);
        let count = {};
        let fePoints = []
        for (let i = 0; i < contours.size(); i++) {
            let k = i;
            let ic = 0;
            while (hierarchy.intPtr(0, k)[2] !== -1) {
                k = hierarchy.intPtr(0, k)[2];
                ic += 1;
            }
            if (ic === 4 && this.isContourCircle(contours.get(i))) {
                k = i;
                k = hierarchy.intPtr(0, k)[3];
                if (!(k in count)) {
                    count[k] = 0;
                }
                count[k]++;
                // 坐标
                let p = cv.moments(contours.get(i))
                fePoints.push({
                    x: p.m10 / p.m00,
                    y: p.m01 / p.m00
                })
                // cv.drawContours(mat, contours, i, [255, 0, 0,255], 2);
            }
        }

        let center = -1;
        for (let obj in count) {
            if (count[obj] === 4) {
                center = parseInt(obj);
                break;
            }
        }
        gray.delete()
        hierarchy.delete()
        if (center === -1) {
            contours.delete()
            return [fePoints,[]];
        }
        // 进行多边形拟合
        let approxContour = new cv.Mat();
        // Initialize vertices
        cv.approxPolyDP(contours.get(center), approxContour, cv.arcLength(contours.get(center), true) * 0.02, true);
        contours.delete()
        if (approxContour.rows !== 4) {
            return [fePoints,[]];
        }
        let points2 = [];
        for (let i = 0; i < approxContour.rows; ++i) {
            let point = approxContour.data32S.slice(i * 2, (i + 1) * 2);
            points2.push({ x: point[0], y: point[1] });
        }
        return [fePoints,points2];
    }
    detectContour(mat) {
        // Convert to grayscale
        let gray = new cv.Mat();
        cv.cvtColor(mat, gray, cv.COLOR_BGR2GRAY);
        cv.GaussianBlur(gray, gray, new cv.Size(3, 3), 0, 0);
        cv.adaptiveThreshold(gray, gray, 255, cv.ADAPTIVE_THRESH_MEAN_C, cv.THRESH_BINARY, 11, 2);
        let contours = new cv.MatVector();
        let hierarchy = new cv.Mat();
        // Find contours
        cv.findContours(gray, contours, hierarchy, cv.RETR_TREE, cv.CHAIN_APPROX_SIMPLE);
        let count = {};
        for (let i = 0; i < contours.size(); i++) {
            let k = i;
            let ic = 0;
            while (hierarchy.intPtr(0, k)[2] !== -1) {
                k = hierarchy.intPtr(0, k)[2];
                ic += 1;
            }
            if (ic === 4 && this.isContourCircle(contours.get(i))) {
                k = i;
                k = hierarchy.intPtr(0, k)[3];
                if (!(k in count)) {
                    count[k] = 0;
                }
                count[k]++;
                // cv.drawContours(mat, contours, i, [255, 0, 0,255], 2);
            }
        }
        // console.log("count",count)
        hierarchy.delete()
        let center = -1;
        for (let obj in count) {
            if (count[obj] === 4) {
                center = parseInt(obj);
                break;
            }
        }
        if (center === -1) {
            contours.delete()
            return [];
        }
        // 进行多边形拟合
        let approxContour = new cv.Mat();
        // Initialize vertices
        cv.approxPolyDP(contours.get(center), approxContour, cv.arcLength(contours.get(center), true) * 0.02, true);
        contours.delete()
        if (approxContour.rows !== 4) {
            return [];
        }
        let points2 = [];
        for (let i = 0; i < approxContour.rows; ++i) {
            let point = approxContour.data32S.slice(i * 2, (i + 1) * 2);
            points2.push({ x: point[0], y: point[1] });
        }
        return points2;
    }

    static processMat(mat) {
        let res = {
            status: 400,
            message: "",
            data: {
                phone: "",
                answer: new Map(),
            },
            points:[],
            pic: []
        };
        // console.log(mat.size())
        let obj = new AnswerSheetDetect();
        let points = [];
        let fePoints = [];
        res.pic = mat
        try {
            [fePoints,points] = obj.getScreenContour(mat);
        }
        catch (e) {
            res.message = "getScreenContour:" + e.toString();
            return res;
        }
        res.points = fePoints
        if (points.length !== 4) {
            res.message = "Screen contour not detected, retrying:" + Date.now().toString();
            return res;
        }
        let innerPoints
        let warped
        try{
            [innerPoints, warped] = obj.getInnerPoint(mat, points);
        }catch (e) {
            res.message = "points or warped err:" + e.toString();
            return res;
        }
        if (!warped || innerPoints.length !== 45) {
            res.message = "Invalid inner points or warped frame, retrying,innerPoints count:" + innerPoints.length;
            res.pic = warped;
            return res;
        }
        let centers
        let sortedGroups
        try{
            // 获取坐标并且按 y 坐标排序
            centers = obj.getCenterSortY(innerPoints);
            // 5 个一组 分组 然后组内按 x 坐标排序 然后合并成 group
            sortedGroups = obj.groupAndSort(centers);
        }catch (e) {
            res.message = "sort err:" + e.toString();
            res.pic = warped;
            return res;
        }
        // 获取手机号
        let phone;
        try {
            phone = obj.getPhone(sortedGroups, warped);
        }
        catch (e) {
            res.message = "getPhone:" + e.toString();
            return res;
        }
        // console.log(phone)
        // if (!(obj.validatePhoneNumber(phone))) {
        //     res.message = "手机号识别失败:" + phone;
        //     return res;
        // }
        // let data = {"phone":phone}
        res.data.phone = phone;
        let ans;
        try {
            ans = obj.getAnswer(sortedGroups.slice(5, sortedGroups.length), warped);
            // console.log(ans)
        }
        catch (e) {
            res.message = "getAnswer:" + e.toString();
            console.log(e.toString());
            return res;
        }
        res.data.answer = ans;
        res.status = 200;
        res.pic = warped;
        return res;
    }
    static processFile(e) {
        return __awaiter(this, void 0, void 0, function* () {
            try{
                const fileInput = e.target || e;
                const file = fileInput.files ? fileInput.files : e;
                if (file) {
                    let fileReader = new FileReader();
                    let imageData = yield new Promise(r => {
                        fileReader.onload = r;
                        fileReader.readAsDataURL(file);
                    });
                    let img = new Image();
                    img.src = imageData.target.result;
                    yield new Promise(r => {
                        img.onload = r;
                    });
                    let mat = cv.imread(img);
                    let result = yield AnswerSheetDetect.processMat(mat);
                    if (result.status !== 200) {
                        return {
                            status: "error",
                            message: result.message,
                        }
                    }
                    // deal ans
                    let arrAns = [];
                    for (let [key, value] of result.data.answer) {
                        arrAns.push({ id: key, answer: value.map(str => str.toLowerCase()) });
                    }
                    return {
                        status: "success",
                        strp: 4,
                        phone: result.data.phone,
                        answer: arrAns,
                    };
                }
            }catch (e) {
                return {
                    status: "error",
                    message: "process file err" + e.toString()
                };
            }
        });
    }
}
export default AnswerSheetDetect;
