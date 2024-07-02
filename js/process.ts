// @ts-nocheck


interface DetectResult {
    status: 'success' | 'error'
    message: string | undefined
    data: {
        phone: string,
        image_list: {
            id: string,
            image: string
        }
    }
}

class Detect {
    private readonly dictionary: any
    private grayList: any[] = []
    private adaptiveList: any[] = []
    private originList: any[] = []
    private idsIndex = 0
    private ids = -1
    private idsCount = 0
    private idsPoint = {}
    private idsMinX = 0
    private idsMinY = 0
    private idsMaxX = 0
    private idsMaxY = 0
    private ZOOM_SCALAR = 1
    private SL_MIN_WIDTH = 1300
    private SL_MAX_WIDTH = 1350
    private SL_MIN_HEIGHT = 10
    private SL_MAX_HEIGHT = 40
    public static tempGray = null
    public static xingce_value = 0.5


    constructor() {
        this.dictionary = new cv.aruco_Dictionary(cv.DICT_4X4_50);
    }

    static resize(input, width, height) {
        let output = new cv.Mat()
        // console.log(input, width, height)
        cv.resize(input, output, new cv.Size(width, height))
        return output
    }


    private static fourPointsTransform(mat, points) {
        let br = points.reduce((prev, current) => (prev.x + prev.y > current.x + current.y) ? prev : current)
        let tl = points.reduce((prev, current) => (prev.x + prev.y < current.x + current.y) ? prev : current)
        let other = points.filter(obj => (obj.x !== br.x || obj.y !== br.y) && (obj.x !== tl.x || obj.y !== tl.y))
        let bl = other.reduce((prev, current) => (prev.y > current.y) ? prev : current)
        let tr = other.reduce((prev, current) => (prev.x > current.x) ? prev : current)
        let widthA = Math.sqrt(Math.pow((br.x - bl.x), 2.0) + Math.pow((br.y - bl.y), 2.0))
        let widthB = Math.sqrt(Math.pow((tr.x - tl.x), 2.0) + Math.pow((tr.y - tl.y), 2.0))
        let heightA = Math.sqrt(Math.pow((tr.x - br.x), 2.0) + Math.pow((tr.y - br.y), 2.0))
        let heightB = Math.sqrt(Math.pow((tl.x - bl.x), 2.0) + Math.pow((tl.y - bl.y), 2.0))
        let maxHeight = heightA > heightB ? heightA : heightB
        let maxWidth = widthA > widthB ? widthA : widthB
        let srcTri = cv.matFromArray(4, 1, cv.CV_32FC2, [tl.x, tl.y, tr.x, tr.y, br.x, br.y, bl.x, bl.y]);
        let dstTri = cv.matFromArray(4, 1, cv.CV_32FC2, [0, 0, maxWidth - 1, 0, maxWidth - 1, maxHeight - 1, 0, maxHeight]);
        let M = cv.getPerspectiveTransform(srcTri, dstTri);
        let four = new cv.Mat()

        let dsize = new cv.Size(maxWidth, maxHeight);
        cv.warpPerspective(mat, four, M, dsize, cv.INTER_LINEAR, cv.BORDER_CONSTANT);
        M.delete()

        return four
    }

    private static euclideanDistance(p1, p2) {
        const dx = p1.x - p2.x;
        const dy = p1.y - p2.y;
        return Math.sqrt(dx * dx + dy * dy);
    }


    private detectMarkersInRects(mat: cv.Mat, rects: cv.Rect[]) {
        const markerIds = new cv.Mat();
        const markerCorners = new cv.MatVector();

        try {
            let as = 0
            for (const rect of rects) {
                const roi = mat.roi(rect);
                // cv.adaptiveThreshold(roi, roi, 255, cv.ADAPTIVE_THRESH_MEAN_C, cv.THRESH_BINARY, 5, 1)
                cv.GaussianBlur(roi, roi, new cv.Size(11, 11), 0);
                // let canv = document.createElement('canvas');
                // document.body.appendChild(canv)
                cv.detectMarkers(roi, this.dictionary, markerCorners, markerIds);
                roi.delete();

                if (markerCorners.size() !== 0) {
                    const corners = markerCorners.get(0);
                    const xData = corners.data32F.filter((_, i) => i % 2 === 0);
                    const yData = corners.data32F.filter((_, i) => i % 2 !== 0);

                    const minX = Math.min(...xData) + rect.x;
                    const maxX = Math.max(...xData) + rect.x;
                    const minY = Math.min(...yData) + rect.y;
                    const maxY = Math.max(...yData) + rect.y;

                    // 返回真实坐标
                    return {
                        id: markerIds.data32S[0],
                        x: minX,
                        y: minY,
                        width: maxX - minX,
                        height: maxY - minY,
                    };
                }
            }
        } finally {
            markerIds.delete();
            markerCorners.delete();
        }

        return null;
    }

    private preprocess(imageList: any[]) {
        let i = 0
        for (let img of imageList) {
            let gray = new cv.Mat()
            let adaptive = new cv.Mat()
            cv.cvtColor(img, gray, cv.COLOR_RGBA2GRAY, 0)
            cv.adaptiveThreshold(gray, adaptive, 255, cv.ADAPTIVE_THRESH_MEAN_C, cv.THRESH_BINARY, 155, 1)
            this.grayList.push(gray)
            this.adaptiveList.push(adaptive)
            // let markerIds = new cv.Mat();
            // let markerCorners = new cv.MatVector();
            let width = img.cols
            let height = img.rows

            // cv.imshow("outCanvas",adaptive)
            let result = this.detectMarkersInRects(gray, [
                new cv.Rect(width - 130, 530, 130, 130),
                new cv.Rect(0, height - 660, 130, 130),

                new cv.Rect(500, 0, 200, 200),
                new cv.Rect(width - 200, 500, 200, 200),
                new cv.Rect(0, 500, 200, 200),
                new cv.Rect(width - 500 - 200, height - 200, 200, 200),

            ])
            // let sp1 = new cv.Point(result.x, result.y)
            // let sp2 = new cv.Point(result.x + result.width, result.y + result.height)
            // cv.rectangle(img, sp1, sp2, [0, 0, 255, 255], 2)
            // cv.imshow("outCanvas",img)
            console.log("result",result)

            if (result) {
                this.idsIndex = i
                this.ids = result.id
                this.idsCount += 1
                this.idsPoint = {x: result.x, y: result.y}
                this.idsMaxX = result.x + result.width
                this.idsMaxY = result.y + result.height
                this.idsMinX = result.x
                this.idsMinY = result.y
            }

            // let tmpMat = new cv.Mat()
            // cv.resize(gray, tmpMat, new cv.Size(Math.floor(gray.cols / this.ZOOM_SCALAR), Math.floor(gray.rows / this.ZOOM_SCALAR)))
            // cv.detectMarkers(tmpMat, this.dictionary, markerCorners, markerIds);
            // tmpMat.delete()
            //
            // if (markerCorners.size() !== 0) {
            //     let corners = markerCorners.get(0)
            //     this.idsIndex = i
            //     this.ids = markerIds.data32S[0]
            //     console.log(markerIds.data32S)
            //     this.idsCount += 1
            //     this.idsPoint = {x: corners.floatAt(0, 0) * this.ZOOM_SCALAR, y: corners.floatAt(0, 1) * this.ZOOM_SCALAR}
            //     let idsX = [corners.floatAt(0, 0), corners.floatAt(0, 2), corners.floatAt(0, 4), corners.floatAt(0, 6)]
            //     let idsY = [corners.floatAt(0, 1), corners.floatAt(0, 3), corners.floatAt(0, 5), corners.floatAt(0, 7)]
            //
            //     this.idsMaxX = Math.max(...idsX) * this.ZOOM_SCALAR
            //     this.idsMaxY = Math.max(...idsY) * this.ZOOM_SCALAR
            //     this.idsMinX = Math.min(...idsX) * this.ZOOM_SCALAR
            //     this.idsMinY = Math.min(...idsY) * this.ZOOM_SCALAR
            //     corners.delete()
            // }
            // markerIds.delete()
            // markerCorners.delete()
            i += 1
        }
    }

    // 计算矩形范围内，黑色像素占比
    private static calcRectWeight(mat, rect) {
        let count = 0
        for (let x = rect.x; x < rect.x + rect.width; x++) {
            for (let y = rect.y; y < rect.y + rect.height; y++) {
                if (mat.ucharAt(y, x) == 0) {
                    count++
                }
            }
        }

        return count / (rect.width * rect.height)
    }

    // 计算手机号函数，根据不同的phoneOption
    private getPhone(phoneOption) {
        let phone = ''
        let start = {x: phoneOption.rect.x, y: phoneOption.rect.y}
        let step = {x: phoneOption.rect.width / 11, y: phoneOption.rect.height / 10}

        for (let x = 0; x < 11; x++) {
            let score = []
            for (let y = 0; y < 10; y++) {
                let currentX = start.x + step.x * x
                let currentY = start.y + step.y * y

                let rect = new cv.Rect(currentX + 11, currentY + 11, 30, 20)
                let sp1 = new cv.Point(rect.x, rect.y)
                let sp2 = new cv.Point(rect.x + rect.width, rect.y + rect.height)
                let weight = Detect.calcRectWeight(this.adaptiveList[this.idsIndex], rect)
                score.push(weight)
                if (weight > 0.4) {
                    cv.rectangle(this.originList[this.idsIndex], sp1, sp2, [255, 0, 0, 255], 2)
                }
                // console.log()
            }

            console.log(score)
            let maxScore = Math.max(...score)
            if (maxScore < 0.4) {
                phone += 'u'
            } else {
                phone += (score.indexOf(maxScore)).toString()
            }
        }
        // cv.imshow('output', this.originList[this.idsIndex])
        return phone
    }


    private static getPhoneFromCamara(originImage, adaptiveImage, phoneOption) {
        // phoneOption: {
        //     x: 503,
        //     y: 92,
        //     width: 485,
        //     height: 230
        // }
        let phone = ''
        let start = {x: phoneOption.rect.x, y: phoneOption.rect.y}
        let step = {x: phoneOption.rect.width / 11, y: phoneOption.rect.height / 10.2}

        for (let x = 0; x < 11; x++) {
            let score = []
            for (let y = 0; y < 10; y++) {
                let currentX = start.x + step.x * x
                let currentY = start.y + step.y * y
                let rect = new cv.Rect(currentX + 11, currentY + 11, 30, 13)
                let sp1 = new cv.Point(rect.x, rect.y)
                let sp2 = new cv.Point(rect.x + rect.width, rect.y + rect.height)
                let weight = this.calcRectWeight(adaptiveImage, rect)
                cv.rectangle(originImage, sp1, sp2, [255, 0, 0, 255], 2)
                score.push(weight)
            }

            let maxScore = Math.max(...score)
            if (maxScore < 0.4) {
                phone += 'u'
            } else {
                phone += (score.indexOf(maxScore)).toString()
            }
        }

        return phone
    }

    private toHsv(mat) {
        let quesImage = mat
        let bgr = new cv.Mat()
        cv.cvtColor(quesImage, bgr, cv.COLOR_RGBA2BGR)
        let hsv = new cv.Mat()
        cv.cvtColor(bgr, hsv, cv.COLOR_BGR2HSV)
        let hsvOutput = new cv.Mat()


        let low = new cv.Mat(hsv.rows, hsv.cols, hsv.type(), [0, 0, 0, 0]);
        let high = new cv.Mat(hsv.rows, hsv.cols, hsv.type(), [180, 255, 80, 0])
        cv.inRange(hsv, low,  high, hsvOutput)
        let canv = document.createElement('canvas');
        // cv.imshow(canv, hsvOutput)
        // document.body.appendChild(canv)

        mat.delete()
        low.delete()
        high.delete()
        hsv.delete()
        hsvOutput.delete()
        bgr.delete()

        return canv.toDataURL()
    }

    static toHsv(mat) {
        let quesImage = mat
        let bgr = new cv.Mat()
        cv.cvtColor(quesImage, bgr, cv.COLOR_RGBA2BGR)
        let hsv = new cv.Mat()
        cv.cvtColor(bgr, hsv, cv.COLOR_BGR2HSV)
        let hsvOutput = new cv.Mat()


        let low = new cv.Mat(hsv.rows, hsv.cols, hsv.type(), [0, 0, 0, 0]);
        let high = new cv.Mat(hsv.rows, hsv.cols, hsv.type(), [180, 255, 80, 0])
        cv.inRange(hsv, low,  high, hsvOutput)
        let canv = document.createElement('canvas');
        cv.imshow(canv, bgr)
        // document.body.appendChild(canv)

        quesImage.delete()
        low.delete()
        high.delete()
        hsv.delete()
        hsvOutput.delete()
        bgr.delete()

        return canv.toDataURL('image/jpeg', 0.6)
    }

    // 处理申论答题卡
    private detectSlCard() {
        let image_list = []
        let width = this.originList[this.idsIndex].cols
        let height = this.originList[this.idsIndex].rows

        for (let i = 0; i < this.originList.length; i++) {
            if ((i === this.idsIndex && this.idsPoint.x > 500) || (this.idsPoint.x <= 500 && i !== this.idsIndex)) {
                cv.rotate(this.originList[i], this.originList[i], cv.ROTATE_90_COUNTERCLOCKWISE)
                cv.rotate(this.grayList[i], this.grayList[i], cv.ROTATE_90_COUNTERCLOCKWISE)
                cv.rotate(this.adaptiveList[i], this.adaptiveList[i], cv.ROTATE_90_COUNTERCLOCKWISE)
            } else {
                cv.rotate(this.originList[i], this.originList[i], cv.ROTATE_90_CLOCKWISE)
                cv.rotate(this.grayList[i], this.grayList[i], cv.ROTATE_90_CLOCKWISE)
                cv.rotate(this.adaptiveList[i], this.adaptiveList[i], cv.ROTATE_90_CLOCKWISE)
            }
        }

        if (this.idsPoint.x > 500) {
            this.idsPoint.x = this.idsMinY
            this.idsPoint.y = width - this.idsMaxX
        } else {
            this.idsPoint.x = height - this.idsMaxY
            this.idsPoint.y = this.idsMinX
        }

        console.log('ggggggggg')

        let rect = new cv.Rect(this.idsPoint.x + 386, this.idsPoint.y + 80, 580, 310)
        let sp1 = new cv.Point(rect.x, rect.y)
        let sp2 = new cv.Point(rect.x + rect.width, rect.y + rect.height)
        cv.rectangle(this.originList[this.idsIndex], sp1, sp2, [0, 0, 255, 255], 2)
        let phone = this.getPhone({
            rect,
            space: {x: 11, y: 11},
            size: {x: 30, y: 22},
        })
        let quesIndex = 1

        for (let i = 0; i < this.originList.length; i++) {
            let corners = new cv.MatVector();
            let nodes = new cv.Mat();
            let tmpInvMat = new cv.Mat()

            cv.morphologyEx(this.adaptiveList[i], this.adaptiveList[i], cv.MORPH_CLOSE, cv.Mat.ones(1, 30, cv.CV_8U));
            // cv.imshow('output', this.adaptiveList[i])

            cv.bitwise_not(this.adaptiveList[i], tmpInvMat);
            cv.findContours(tmpInvMat, corners, nodes, cv.RETR_LIST, cv.CHAIN_APPROX_SIMPLE)
            tmpInvMat.delete()


            let left_rectangles = []
            let right_rectangles = []

            for (let j = 0; j < corners.size(); j++) {
                let contour = corners.get(j)
                let rect = cv.boundingRect(contour)
                let minAreaRect = cv.minAreaRect(contour)
                contour.delete()

                let realSize = minAreaRect.size
                let realHeight = Math.min(realSize.width, realSize.height)
                let realWidth = Math.max(realSize.width, realSize.height)
                if (this.SL_MIN_WIDTH <= realWidth && realWidth <= this.SL_MAX_WIDTH && realHeight > this.SL_MIN_HEIGHT && realHeight < this.SL_MAX_HEIGHT) {
                    cv.drawContours(this.originList[i], corners, j, [255, 0, 0, 255], 2)
                    if (rect.x < width / 2) {
                        left_rectangles.push({
                            rect,
                            width: realWidth,
                            height: realHeight
                        })
                    } else {
                        right_rectangles.push({
                            rect,
                            width: realWidth,
                            height: realHeight
                        })
                    }
                }
            }
            corners.delete()
            nodes.delete()
            // cv.imshow('output', this.originList[i])

            for (let rectangles of [left_rectangles, right_rectangles]) {
                let [page_w, page_h] = [1450, 2290]
                let pageStart = rectangles.find(obj => obj.height <= 20 && obj.rect.y < 100)
                rectangles.sort((a, b) => a.rect.y - b.rect.y)
                if (pageStart == null) {
                    pageStart = rectangles[0]
                    pageStart.rect.y += 20
                }

                rectangles = rectangles.filter(obj => obj != pageStart && obj.height > 30 && obj.height < 40)
                for (let split of rectangles) {
                    let roiRect = new cv.Rect(pageStart.rect.x, pageStart.rect.y + 20, page_w, split.rect.y - 30 - pageStart.rect.y)
                    console.log(roiRect)
                    let image = this.toHsv(this.originList[i].roi(roiRect))
                    image_list.push({
                        id: quesIndex,
                        image
                    })
                    pageStart = split
                    pageStart.rect.y += 20
                    quesIndex += 1
                }

                let roiRect = new cv.Rect(pageStart.rect.x, pageStart.rect.y + 20, page_w, page_h - 30 - pageStart.rect.y)
                let image = this.toHsv(this.originList[i].roi(roiRect))
                image_list.push({
                    id: quesIndex,
                    image
                })
            }
        }

        return {
            status: "success",
            data: {
                phone,
                image_list
            }
        }
    }

    public static async initTemplate() {
        let tempGray = await Detect.readMatFromBase64('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABwAAAAcCAYAAAByDd+UAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAC9SURBVEhL7Y5JDsQgDAT5/6dJfCACU94yKKcpqQ/BvaT1j/kPLrTWNlUJEzTiKcJ1UGFGHnilkjcitlcKijzILyLCwQo6S/nlJTJn0B265/mKjBW8HnPwF7yu0qD2WD7B8uAgMd+1CMtzZFCkse5HBgnLg4OzYUZ7hgjLUxoU3vhmyoMZvK7lyzNW8HrcQW3OEOW3Fx2gEEE5kQbbKDikIc8QYf4+FVRkYV9uqChSROy4oWJShpzrIB8P9n4BbDAcY+LUK9IAAAAASUVORK5CYII=')
        cv.cvtColor(tempGray, tempGray, cv.COLOR_RGBA2GRAY, 0)
        return tempGray
    }

    public static detectFromCamara(mat, tempGray) {
        let resp = {}
        let points = []
        let grayImg = new cv.Mat()
        let cannyImg = new cv.Mat()
        resp.status = 'success'

        cv.cvtColor(mat, grayImg, cv.COLOR_RGBA2GRAY, 0)
        cv.GaussianBlur(grayImg, grayImg, new cv.Size(3, 3), 0, 0)
        cv.adaptiveThreshold(grayImg, grayImg, 255, cv.ADAPTIVE_THRESH_GAUSSIAN_C, cv.THRESH_BINARY, 15, 1)
        cv.Canny(grayImg, cannyImg, 10, 400)

        let contours = new cv.MatVector();
        let hierarchy = new cv.Mat();
        cv.findContours(grayImg, contours, hierarchy, cv.RETR_TREE, cv.CHAIN_APPROX_SIMPLE);

        try {
            for (let i = 0; i < contours.size(); i++) {
                let k = i
                let ic = 0
                while (hierarchy.intPtr(0, k)[2] !== -1) {
                    k = hierarchy.intPtr(0, k)[2]
                    ic++
                }
                if (ic == 4) {
                    let area = cv.contourArea(contours.get(i))
                    if (area < 1200 || area > 2500){
                        continue
                    }
                    cv.drawContours(mat, contours, i, [255, 0, 0, 255], 2)
                    let p = cv.moments(contours.get(i))
                    points.push({
                        x: p.m10 / p.m00,
                        y: p.m01 / p.m00
                    })
                }
            }

            // cv.imshow("outCanvas",mat)
            // console.log("points",points)

            if (points.length !== 4) {
                console.log("points.length !=== 4:",points.length)
                return resp
            }

            resp.strp = 1

            let four = Detect.fourPointsTransform(mat, points)

            let origin = Detect.resize(four, 248 * 4, 351 * 4)

            four.delete()

            four = new cv.Mat()
            cv.cvtColor(origin, four, cv.COLOR_RGBA2GRAY, 0)
            cv.adaptiveThreshold(four, four, 255, cv.ADAPTIVE_THRESH_MEAN_C, cv.THRESH_BINARY, 11, 2)

            // console.log("four",four)
            resp.phone = Detect.getPhoneFromCamara(origin, four, {
                rect: {
                    x: 503,
                    y: 92,
                    width: 485,
                    height: 230
                }
            })

            console.log("phone",resp.phone)
            let res = new cv.Mat()
            cv.matchTemplate(four, tempGray, res, cv.TM_CCOEFF_NORMED);
            resp.strp = 1.5

            const threshold = 0.7;
            const [w, h] = [tempGray.cols, tempGray.rows];

            const locationsData = res.data32F;
            // console.log(locationsData)
            let locations = [];
            for (let i = 0; i < locationsData.length; i++) {
                if (locationsData[i] >= threshold) {
                    const x = i % res.cols;
                    const y = Math.floor(i / res.cols);
                    locations.push({
                        point: new cv.Point(x + w / 2, y + h / 2),
                        score: locationsData[i]
                    });
                }
            }

            // console.log(locations)
            resp.strp = 2
            locations.sort((a, b) => b.score - a.score);
            // 过滤距离小于 50 的重复点
            let filteredLocations = [];
            for (const loc of locations) {
                const isNearbyHigherScore = filteredLocations.some(filteredLoc => {
                    const dist = Detect.euclideanDistance(loc.point, filteredLoc.point);
                    return dist < 50 && filteredLoc.score >= loc.score;
                });
                if (!isNearbyHigherScore) {
                    filteredLocations.push(loc);
                }
            }
            resp.strp = 3
            filteredLocations = filteredLocations.map(obj => obj.point)
            filteredLocations = filteredLocations.filter(obj => obj .x > 100 && obj.x < 900 && obj.y < 1330 && obj.y > 100)
            filteredLocations = filteredLocations.sort((a, b) => b.x - a.x);

            for (let i = 0; i < filteredLocations.length; i++) {
                cv.circle(origin, filteredLocations[i], 5, [0, 255, 0, 255], 5)
            }

            resp.strp = 4

            if (filteredLocations.length === 21) {
                filteredLocations.sort((a, b) => b.y - a.y)
                filteredLocations.reverse()
                let rowList = []

                let index = 0
                rowList[0] = []
                rowList[0].push(filteredLocations[0])
                // console.log(filteredLocations[0])
                filteredLocations[0].index = 0
                for (let i = 1; i < filteredLocations.length; i++) {
                    filteredLocations[i].index = i;
                    if (filteredLocations[i].y - filteredLocations[i - 1].y > 50) {
                        index++
                        rowList[index] = []
                    }
                    rowList[index].push(filteredLocations[i])
                }

                for (let row of rowList) {
                    const avgY = row.reduce((sum, point) => sum + point.y, 0) / row.length;
                    // 添加第一个点,x坐标为5,y坐标为平均值
                    const point1 = new cv.Point(5, avgY);
                    row.unshift(point1);
                    // 添加第二个点,x坐标为four的宽度-5,y坐标和第一个点相同
                    const point2 = new cv.Point(four.cols - 5, avgY);
                    row.push(point2);
                    row.sort((a, b) => a.x - b.x);
                }

                // 添加最后一行点
                const lastRow = rowList[0].map(point => new cv.Point(point.x, four.rows - 5));
                rowList.push(lastRow);

                let cardList = []
                // let shows = []
                for (let i = 0; i < rowList.length - 1; i++) {
                    const currentRow = rowList[i];
                    const nextRow = rowList[i + 1];

                    for (let j = 0; j < 4; j++) {
                        const rect = [
                            currentRow[j],
                            currentRow[j + 1],
                            nextRow[j],
                            nextRow[j + 1]
                        ];

                        const rectImg = Detect.fourPointsTransform(origin, rect);


                        let canv = document.createElement('canvas');
                        let cardMat = new cv.Mat()
                        let cardGrayMat = new cv.Mat()
                        cv.cvtColor(rectImg, cardGrayMat, cv.COLOR_RGBA2GRAY, 0)
                        cv.adaptiveThreshold(cardGrayMat, cardMat, 255, cv.ADAPTIVE_THRESH_MEAN_C, cv.THRESH_BINARY, 33, 1)
                        let cardMatResized = Detect.resize(cardMat, 250, 150)

                        // let rr = Detect.resize(rectImg, 250, 150)
                        // shows.push(rr);
                        cardList.push(cardMatResized)
                        // cv.imshow(canv, cardMatResized)
                        document.body.appendChild(canv)
                        rectImg.delete()
                        cardGrayMat.delete()
                        cardMat.delete()
                    }
                }
                let ns = 0
                resp.answer = []

                let tmIndex = 0
                for (let card of cardList) {
                    // let last = Math.floor(card.rows / 5);
                    // 定义感兴趣的区域（ROI）
                    // let rect = new cv.Rect(0, last, card.cols, card.rows - last);
                    // let dst = card.roi(rect)
                    // let showa = shows[ns]
                    // let showaaa = showa.roi(rect)
                    // let ans = Detect.dealAnswer(dst,showaaa,ns)
                    // console.log(ans)
                    // ns++

                    let defaultStartX = 19 + (ns % 5)
                    let defaultStartY = 41
                    if( Math.floor(ns / 5) == 4){
                        defaultStartY += 4
                    }

                    for (let x = 0; x < 5; x++) {
                        tmIndex += 1
                        let ans = []
                        for (let y = 0; y < 4; y++) {
                            let startX = defaultStartX + x * 44
                            let startY = defaultStartY + y * 24
                            let w = 30
                            let h = 15

                            // let sp1 = new cv.Point(startX, startY)
                            // let sp2 = new cv.Point(startX + w, startY + h)
                            // cv.rectangle(shows[ns], sp1, sp2, [0, 0, 255, 255], 2)

                            // console.log(this.xingce_value, '0000')
                            // console.log(Detect.calcRectWeight(card, new cv.Rect(startX, startY, w, h)))
                            if (Detect.calcRectWeight(card, new cv.Rect(startX, startY, w, h)) > this.xingce_value) {
                                ans.push('abcd'[y])
                            }
                        }
                        // console.log("------")
                        // console.log(ans)

                        resp.answer.push({
                            "id": tmIndex,
                            "answer": ans
                        })
                    }
                    ns++
                }

                // for(let i = 0;i <shows.length;i++){
                //     let nn = "ccccc" + i
                //     cv.imshow(nn,shows[i])
                // }
                // cv.imshow("outCanvas",origin)
                // console.log(cardList)

            } else {
                resp.status = 'error'
                resp.message = '区域缺失，请确保没有处于强光/弱光状态，或者尽量保持答题卡纸面平整'
            }

            origin.delete()
            four.delete()
            res.delete()
            console.log("respppp",resp)
        } catch (e) {
            console.error(e, e.stack)
            resp.status = 'error'
            resp.message = e.toString()
            return resp
        } finally {
            hierarchy.delete()
            contours.delete()
            grayImg.delete()
            cannyImg.delete()
        }

        return resp
    }
    static dealAnswer(dst,showaaa,ns){
        let contours = new cv.MatVector();
        let hierarchy = new cv.Mat();
        // Find contours
        cv.findContours(dst, contours, hierarchy, cv.RETR_TREE, cv.CHAIN_APPROX_SIMPLE);
        let all_center = []
        let center2contour = new Map()
        let allW = 0
        let allH = 0
        for (let i = 0; i < contours.size(); i ++) {
            let contour = contours.get(i)

            let minAreaRect = cv.minAreaRect(contour)
            // console.log("ddd",minAreaRect.size.width,minAreaRect.size.height)
            if (minAreaRect.size.width > dst.size().width / 6 || minAreaRect.size.height > dst.size().height / 4.5) {
                continue
            }

            if (minAreaRect.size.width < dst.size().width / 8 || minAreaRect.size.height < dst.size().height / 6.6) {
                continue
            }
            allW += minAreaRect.size.width
            allH += minAreaRect.size.height
            all_center.push(minAreaRect.center)
            center2contour.set(minAreaRect.center,contour)
            cv.drawContours(dst, contours, i, [255, 0, 0,255], 1);
        }
        if (all_center.length == 0 ) {
            return []
        }
        // console.log(all_center)
        // 轮廓去重
        let uniqueCenters = Detect.removeSimilarCoordinates(all_center, 5)
        // console.log(uniqueCenters)
        let completeCenters = Detect.completeAnswer(uniqueCenters,dst.size().width,dst.size().height)
        // console.log(completeCenters)
        let defaultWeight = Math.round(allW / center2contour.size * 10) / 10
        let defaultHeight = Math.round(allH / center2contour.size * 10) / 10
        // console.log(defaultWeight, defaultHeight)
        let resAnswer = []
        // resAnswer.push(ansSelect)
        for (let groupIndex = 0; groupIndex < completeCenters.length; groupIndex++) {
            let group = completeCenters[groupIndex]
            let tmpBack = []
            let groupAnsList = []
            for (let j = 0; j < group.length; j++) {
                let contour = null
                if (center2contour.has(group[j])){
                    contour = center2contour.get(group[j])
                }else{
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
                }

                // let l = new cv.Mat()
                // cv.approxPolyDP(contour,l,0.025 * cv.arcLength(contour, true),true)
                let rect = cv.boundingRect(contour)

                // let sp1 = new cv.Point(rect.x, rect.y)
                // let sp2 = new cv.Point(rect.x + rect.width, rect.y + rect.height)
                // cv.rectangle(showaaa, sp1, sp2, [0, 0, 255, 255], 1)

                // console.log("rect",rect.x,rect.y,rect.width,rect.height)
                let w = Math.round(rect.width / 8 * 10) / 10;
                let h = Math.round(rect.height / 7 * 10) / 10;
                rect.x = rect.x + w
                rect.y = rect.y + h
                rect.width = rect.width - w * 1.8
                rect.height = rect.height - h * 1.8
                // console.log("after",rect.x, rect.y, rect.width, rect.height)
                let optionsImage = dst.roi(rect)
                // cv.rectangle(showaaa, sp1, sp2, [0, 0, 255, 255], 1)
                let resizeImage = Detect.resize(optionsImage,defaultWeight,defaultHeight)
                // imageArr.push(resizeImage)
                let blackPixelCount = Detect.countBlack(resizeImage)
                // let total = resizeImage.width * resizeImage.height
                let total = defaultWeight*defaultHeight
                // console.log('blackPixelCount', Math.round(blackPixelCount/total * 10000) / 10000)
                if (Math.round(blackPixelCount/total * 10000) / 10000 > 0.4) {
                    tmpBack.push(Detect.getOptionsAns(j))
                }
            }
            resAnswer.push(tmpBack)
            // console.log(tmpBack)
            // console.log("-------")
        }
        // return dst
        // return [resAnswer,imageArr]
        // let n = "ccccc" + ns
        // cv.imshow(n,showaaa)
        return resAnswer
    }
    static getOptionsAns(j) {
        switch (j){
            case 0:
                return "a"
            case 1:
                return "b"
            case 2:
                return "c"
            case 3:
                return "d"
        }
    }
    static countBlack(image){
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
    static completeAnswer(uniqueCenters,dstW,dstH){
        let default_y = dstW * 0.078
        let default_x = dstH * 0.29
        uniqueCenters.sort((a, b) => a.y - b.y);
        if (uniqueCenters[0].y < 45){
            let group = Detect.findSimilarY(uniqueCenters,uniqueCenters[0].y)
            let all = 0
            for (let i = 0; i < group.length; i++) {
                all += group[i].y
            }
            default_y = Math.round(all/group.length * 10)/10
        }
        uniqueCenters.sort((a, b) => a.x - b.x);
        if (uniqueCenters[0].x < 70){
            let group = Detect.findSimilarX(uniqueCenters,uniqueCenters[0].x)
            let all = 0
            for (let i = 0; i < group.length; i++) {
                all += group[i].x
            }
            default_x = Math.round(all/group.length * 10)/10
        }
        // console.log(default_x,default_y)
        // 圆心直接的平均宽度距离
        let avg_w = Detect.getAvgW(uniqueCenters,dstW/5,10)
        // 圆心直接的平均高度距离
        let avg_h = Detect.getAvgH(uniqueCenters,10,dstH/4.5)
        // console.log("aaaa",default_x,default_y,avg_w,avg_h)
        uniqueCenters.sort((a, b) => a.y - b.y);
        let res = []

        for (let i = 0;i < 5;i++){
            let findX = default_x + avg_w * i
            let groupX = Detect.findSimilarX(uniqueCenters,findX)
            if (groupX.length == 0 ){
                groupX.push({x:findX,y:default_y})
            }
            if (groupX.length == 4){
                res.push(groupX)
                continue
            }
            groupX.sort((a, b) => a.y - b.y)
            let pre = groupX[0]
            let step = 0

            if (!(Detect.isSimilar(pre,{x:findX,y:default_y},5))) {
                // 计算按 y 坐标排序后的第一位,是不是第一个元素,不是的话按 num 补全curr前面的元素
                step = Math.round(Math.abs(pre.y - default_y) / avg_h)
                for (let z = 1; z <= step; z++) {
                    groupX.unshift({x:pre.x,y:pre.y - z * avg_h})
                }
            }
            if (groupX.length == 4){
                res.push(groupX)
                continue
            }
            let curr = step
            while (curr < 3){
                if (curr == groupX.length - 1){
                    let currItem = groupX[curr]
                    for (let z = 1; z < 4-curr; z++) {
                        groupX.push({x:currItem.x,y:currItem.y + z * avg_h})
                    }
                    break
                }else{
                    let next = curr + 1
                    let nextItem = groupX[next]
                    let currItem = groupX[curr]

                    let num = Math.round(Math.abs(nextItem.y - currItem.y) / avg_h)
                    if (num > 1){
                        let preGroup = groupX.slice(0,curr+1)
                        let lastGroup = groupX.slice(next,groupX.length)
                        let tmpGroup = []
                        for (let j = 1; j < num; j++){
                            tmpGroup.push({x:currItem.x,y:currItem.y + j * avg_h})
                        }
                        groupX = [...preGroup,...tmpGroup,...lastGroup]
                    }
                    curr = curr + num
                }
            }
            res.push(groupX)
        }
        return res
    }

    static isSimilar(coord1, coord2, threshold) {
        return Math.abs(coord1.x - coord2.x) < threshold && Math.abs(coord1.y - coord2.y) < threshold;
    }
    static getAvgW(uniqueCenters,diffX,diffY){
        uniqueCenters.sort((a, b) => a.y - b.y);
        let newGroup = []
        let curr = 0;
        let i = 1;
        while (i < uniqueCenters.length) {
            let tmp = []
            tmp.push(uniqueCenters[curr]);
            // while (i < uniqueCenters.length && Math.abs(uniqueCenters[curr].y - uniqueCenters[i].y) < 10){
            while (i < uniqueCenters.length && Math.abs(uniqueCenters[curr].y - uniqueCenters[i].y) < diffY){
                tmp.push(uniqueCenters[i]);
                i++
            }
            newGroup.push(tmp);
            curr = i
            i++
        }
        let allW = 0
        let wCount = 0
        for (let i = 0; i < newGroup.length; i++) {
            let tmpGroup = newGroup[i];
            tmpGroup.sort((a, b) => a.x - b.x);
            if (tmpGroup.length < 2 ){
                continue
            }
            let pre = 0
            let curr = 1
            while (curr < tmpGroup.length){
                // if (Math.abs(tmpGroup[pre].x - tmpGroup[curr].x) < 30){
                if (Math.abs(tmpGroup[pre].x - tmpGroup[curr].x) < diffX){
                    allW += Math.abs(tmpGroup[pre].x - tmpGroup[curr].x)
                    wCount++
                }
                pre++
                curr++
            }
        }
        return Math.round(allW/wCount * 10) / 10
    }
    static getAvgH(uniqueCenters,diffX,diffY){
        uniqueCenters.sort((a, b) => a.x - b.x);

        let newGroup = []
        let curr = 0;
        let i = 1;
        while (i < uniqueCenters.length) {
            let tmp = []
            tmp.push(uniqueCenters[curr]);
            // while (i < uniqueCenters.length && Math.abs(uniqueCenters[curr].x - uniqueCenters[i].x) < 5){
            while (i < uniqueCenters.length && Math.abs(uniqueCenters[curr].x - uniqueCenters[i].x) < diffX){
                tmp.push(uniqueCenters[i]);
                i++
            }
            newGroup.push(tmp);
            curr = i
            i++
        }
        let allH = 0
        let hCount = 0
        for (let i = 0; i < newGroup.length; i++) {
            let tmpGroup = newGroup[i];
            tmpGroup.sort((a, b) => a.y - b.y);
            if (tmpGroup.length < 2 ){
                continue
            }
            let pre = 0
            let curr = 1
            while (curr < tmpGroup.length){
                // if (Math.abs(tmpGroup[pre].y - tmpGroup[curr].y) < 23){
                if (Math.abs(tmpGroup[pre].y - tmpGroup[curr].y) < diffY){
                    allH += Math.abs(tmpGroup[pre].y - tmpGroup[curr].y)
                    hCount++
                }
                pre++
                curr++
            }
        }
        return  Math.round(allH/hCount * 10) / 10
    }

    private detectXcCard() {
        let width = this.originList[this.idsIndex].cols
        let height = this.originList[this.idsIndex].rows

        if (width > height) {
            if (this.idsPoint.x > 500) {
                this.idsPoint.x = this.idsMinY
                this.idsPoint.y = width - this.idsMaxX
                cv.rotate(this.originList[0], this.originList[0], cv.ROTATE_90_COUNTERCLOCKWISE)
                cv.rotate(this.grayList[0], this.grayList[0], cv.ROTATE_90_COUNTERCLOCKWISE)
                cv.rotate(this.adaptiveList[0], this.adaptiveList[0], cv.ROTATE_90_COUNTERCLOCKWISE)
            } else {
                this.idsPoint.x = height - this.idsMaxY
                this.idsPoint.y = this.idsMinX
                cv.rotate(this.originList[0], this.originList[0], cv.ROTATE_90_CLOCKWISE)
                cv.rotate(this.grayList[0], this.grayList[0], cv.ROTATE_90_CLOCKWISE)
                cv.rotate(this.adaptiveList[0], this.adaptiveList[0], cv.ROTATE_90_CLOCKWISE)
            }
        } else if (this.idsPoint.y > height / 2) {
            // 逆时针旋转180度
            this.idsPoint.x = width - this.idsMinX;
            this.idsPoint.y = height - this.idsMinY;
            cv.rotate(this.originList[0], this.originList[0], cv.ROTATE_180);
            cv.rotate(this.grayList[0], this.grayList[0], cv.ROTATE_180);
            cv.rotate(this.adaptiveList[0], this.adaptiveList[0], cv.ROTATE_180);
        }

        let tmpMat = Detect.resize(this.originList[0], 1080, 1920)
        this.originList[0].delete()
        this.originList[0] = tmpMat

        // console.log(tmpMat)

        return Detect.detectFromCamara(this.originList[0], Detect.tempGray)
    }

    private detectInternal(imageList: any[]) {
        this.originList = imageList

        // 对传入的图片进行预处理
        this.preprocess(imageList)
        if (this.idsCount != 1) {
            return {
                status: 'error',
                message: this.idsCount === 0 ? '未找到标志' : '标志不唯一'
            }
        }

        if (this.idsIndex !== 0) {
            this.grayList.unshift(this.grayList.splice(this.idsIndex, 1)[0])
            this.adaptiveList.unshift(this.adaptiveList.splice(this.idsIndex, 1)[0])
            this.originList.unshift(this.originList.splice(this.idsIndex, 1)[0])
            this.idsIndex = 0
        }

        // 根据检测到的标志，处理类型不同的答题卡
        switch (this.ids) {
            case 0: return this.detectXcCard()
            case 1: return this.detectSlCard()
            default: return {
                status: "error",
                message: "未知标记:" + this.ids
            }
        }
    }

    // 唯一入口
    static detect(imageList: any[]): DetectResult {
        let obj = new Detect()

        try {
            return obj.detectInternal(imageList)
        } catch (e) {
            console.error(e.stack)
            return {
                status: "error",
                message: "处理异常：" + e.toString()
            }
        } finally {
            obj.clear()
        }
    }

    private clear() {
        this.dictionary.delete()
        this.grayList.forEach(obj => obj.delete())
        this.grayList = []

        this.adaptiveList.forEach(obj => obj.delete())
        this.adaptiveList = []

        this.originList.forEach(obj => obj.delete())
        this.originList = []
    }

    static async readMatFromBase64(imgStr: string) {
        let img = new Image()
        img.src = imgStr
        await new Promise(r => {
            img.onload = r
        })

        // @ts-ignore
        return cv.imread(img)
    }
    static removeSimilarCoordinates(all_center, threshold) {
        let uniqueCenters = [];
        for (let center of all_center) {
            let isUnique = true;
            for (let uniqueCenter of uniqueCenters) {
                if (Detect.isSimilar(center, uniqueCenter, threshold)) {
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

    static findSimilarX(uniqueCenters,findX){
        let group = []
        for (let item of uniqueCenters) {
            if (Math.abs(item.x - findX) < 8){
                group.push(item);
            }
        }
        return group;
    }
    static findSimilarY(uniqueCenters,findY){
        let group = []
        for (let item of uniqueCenters) {
            if (Math.abs(item.y - findY) < 10){
                group.push(item);
            }
        }
        return group;
    }

    static async readMatFromFile(e: any, index?: number,type:any) {
        index = index || 0
        const fileInput = e.target || e;
        const file = fileInput.files ? fileInput.files[index] : e;
        if(type=='1'){
            this.xingce_value=0.5
        }
        // console.log(file)
        if (file) {
            let fileReader = new FileReader()

            let imageData = await new Promise(r => {
                fileReader.onload = r
                fileReader.readAsDataURL(file)
            }) as any

            let img = new Image()
            img.src = imageData.target.result
            await new Promise(r => {
                img.onload = r
            })

            // @ts-ignore
            return cv.imread(img)
        }
    }
}



export default Detect
